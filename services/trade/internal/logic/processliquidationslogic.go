package logic

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessLiquidationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProcessLiquidationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessLiquidationsLogic {
	return &ProcessLiquidationsLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ProcessLiquidationsLogic) ProcessPosition(positionID int64) error {
	position, err := l.svcCtx.ContractPositionModel.FindOne(l.ctx, positionID)
	if err != nil {
		return err
	}
	if !position.Qty.IsPositive() || position.MarginMode != int64(trade.MarginMode_MARGIN_MODE_ISOLATED) {
		return nil
	}
	contract, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, position.TenantId, position.SymbolId)
	if err != nil {
		return err
	}
	liquidation, err := l.svcCtx.ContractLiquidationModel.FindActiveByPosition(l.ctx, position.TenantId, position.Id)
	if errors.Is(err, models.ErrNotFound) {
		liqNo := fmt.Sprintf("LIQ-%d-%d", position.Id, position.Version)
		now := utils.NowMillis()
		equity := position.PositionMargin.Add(position.IsolatedMargin).Add(position.UnrealizedPnl)
		if position.MarkSnapshotId == "" {
			return errors.New("liquidation requires confirmed mark price snapshot")
		}
		res, insertErr := l.svcCtx.ContractLiquidationModel.Insert(l.ctx, &models.TContractLiquidation{TenantId: position.TenantId, LiquidationNo: liqNo, PositionId: position.Id, UserId: position.UserId, SymbolId: position.SymbolId, PositionSide: position.PositionSide, MarginMode: position.MarginMode, TriggerMarkPrice: position.MarkPrice, TriggerSnapshotId: position.MarkSnapshotId, TriggerQty: position.Qty, MaintenanceMargin: position.MaintenanceMargin, AccountEquity: equity, BankruptcyPrice: position.BankruptcyPrice, Status: int64(trade.LiquidationStatus_LIQUIDATION_STATUS_PENDING_TAKEOVER), Reason: "isolated maintenance margin breached", CreateTimes: now, UpdateTimes: now})
		if insertErr != nil {
			return insertErr
		}
		id, _ := res.LastInsertId()
		liquidation, err = l.svcCtx.ContractLiquidationModel.FindOne(l.ctx, id)
	}
	if err != nil {
		return err
	}
	if liquidation.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED) {
		return nil
	}
	if err := l.lockRiskUnit(position, liquidation); err != nil {
		return err
	}
	if err := l.cancelRiskIncreasingOrders(position); err != nil {
		return err
	}
	return l.settleTakeover(position, contract, liquidation)
}

func (l *ProcessLiquidationsLogic) lockRiskUnit(position *models.TContractPosition, liq *models.TContractLiquidation) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		pm := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)
		lm := models.NewTContractLiquidationModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, position.TenantId, position.UserId, position.SymbolId, position.PositionSide, position.MarginMode)
		if err != nil {
			return err
		}
		if !current.Qty.IsPositive() {
			return nil
		}
		current.Status = int64(trade.PositionStatus_POSITION_STATUS_LIQUIDATING)
		current.Version++
		current.UpdateTimes = utils.NowMillis()
		if err := pm.Update(ctx, current); err != nil {
			return err
		}
		liq.Status = int64(trade.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING)
		if liq.StartedAt == 0 {
			liq.StartedAt = utils.NowMillis()
		}
		liq.Version++
		liq.UpdateTimes = utils.NowMillis()
		return lm.Update(ctx, liq)
	})
}

func (l *ProcessLiquidationsLogic) cancelRiskIncreasingOrders(position *models.TContractPosition) error {
	cursor := int64(0)
	for {
		statuses := append(matchableOrderStatuses(), int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING))
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{TenantId: position.TenantId, UserId: position.UserId, SymbolId: position.SymbolId, ProductType: int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE), Statuses: statuses}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, o := range orders {
			cursor = o.Id
			if o.IsReduceOnly == int64(common.YesNo_YES_NO_YES) {
				continue
			}
			o.Status = int64(trade.OrderStatus_ORDER_STATUS_CANCELING)
			o.CanceledQty = decimalMaxZero(o.Qty.Sub(o.FilledQty))
			o.CancelReason = "risk liquidation"
			o.Version++
			o.UpdateTimes = utils.NowMillis()
			if err := l.svcCtx.TradeOrderModel.Update(l.ctx, o); err != nil {
				return err
			}
			if err := removeOrderBookOrder(l.svcCtx, l.ctx, o); err != nil {
				return err
			}
			if err := unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, o, "liquidation risk order release"); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessLiquidationsLogic) settleTakeover(position *models.TContractPosition, contract *models.TTradeSymbolContract, liq *models.TContractLiquidation) error {
	pnl := contractRealizedPnl(position.PositionSide, position.OpenAvgPrice, position.MarkPrice, position.Qty, contract.ContractSize, position.ContractValueType)
	values, err := calculateContractTradeValues(position.ContractValueType, position.Qty, contract.ContractSize, position.MarkPrice)
	if err != nil {
		return err
	}
	fee := roundContractDebit(values.SettlementNotional.Mul(contract.LiquidationFeeRate))
	equity := position.PositionMargin.Add(position.IsolatedMargin).Add(pnl).Sub(fee)
	if equity.IsPositive() {
		if err := l.assetChange(position.TenantId, position.UserId, position.MarginAsset, equity, true, liq.Id, liq.LiquidationNo+"-RESIDUAL", "liquidation residual equity"); err != nil {
			return err
		}
	}
	deficit := decimalMaxZero(equity.Neg())
	var fund *models.TContractInsuranceFundAccount
	if deficit.IsPositive() {
		fund, err = l.svcCtx.ContractInsuranceFundModel.FindEnabled(l.ctx, position.TenantId, position.SymbolId, position.MarginAsset)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if err == nil {
			covered, remaining, coverErr := l.tryInsurance(fund, deficit, liq)
			if coverErr != nil {
				return coverErr
			}
			liq.InsuranceFundAmount = covered
			deficit = remaining
		}
	}
	if deficit.IsPositive() {
		if fund == nil || fund.AdlEnabled != int64(common.YesNo_YES_NO_YES) {
			liq.Status = int64(trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW)
			liq.Reason = "insurance fund unavailable or insufficient"
			liq.Version++
			liq.UpdateTimes = utils.NowMillis()
			return l.svcCtx.ContractLiquidationModel.Update(l.ctx, liq)
		}
		adlQty, remaining, err := l.executeADL(position, contract, liq, deficit)
		if err != nil {
			return err
		}
		liq.AdlQty = adlQty
		if remaining.IsPositive() {
			liq.Status = int64(trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW)
			liq.Reason = "ADL liquidity insufficient"
			liq.Version++
			liq.UpdateTimes = utils.NowMillis()
			return l.svcCtx.ContractLiquidationModel.Update(l.ctx, liq)
		}
	}
	return l.completeLiquidation(position, liq, fee)
}

func (l *ProcessLiquidationsLogic) tryInsurance(fund *models.TContractInsuranceFundAccount, amount decimal.Decimal, liq *models.TContractLiquidation) (decimal.Decimal, decimal.Decimal, error) {
	resp, err := l.svcCtx.AssetClient.CoverInsuranceDeficit(l.ctx, &asset.CoverInsuranceDeficitReq{TenantId: fund.TenantId, FundUserId: fund.FundUserId, WalletType: common.WalletType(fund.WalletType), Coin: fund.SettleAsset, RequestedAmount: amount.String(), LiquidationId: liq.Id, LiquidationNo: liq.LiquidationNo + "-INSURANCE", Remark: "liquidation insurance fund cover"})
	if err != nil {
		return decimal.Zero, amount, err
	}
	if resp == nil || resp.GetBase() == nil {
		return decimal.Zero, amount, errors.New("empty insurance fund response")
	}
	if resp.GetBase().GetCode() != 200 {
		return decimal.Zero, amount, fmt.Errorf("insurance fund rejected: %s", resp.GetBase().GetMsg())
	}
	covered, err := decimal.NewFromString(resp.CoveredAmount)
	if err != nil {
		return decimal.Zero, amount, err
	}
	remaining, err := decimal.NewFromString(resp.RemainingAmount)
	if err != nil {
		return decimal.Zero, amount, err
	}
	if covered.IsNegative() || remaining.IsNegative() || !covered.Add(remaining).Equal(amount) {
		return decimal.Zero, amount, errors.New("invalid insurance fund coverage response")
	}
	return covered, remaining, nil
}

func (l *ProcessLiquidationsLogic) assetChange(tenant, user int64, coin string, amount decimal.Decimal, credit bool, bizID int64, bizNo, remark string) error {
	var resp *asset.ChangeAssetResp
	var err error
	if credit {
		resp, err = l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{TenantId: tenant, UserId: user, WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: coin, Amount: amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: bizID, BizNo: bizNo, Remark: remark})
	} else {
		resp, err = l.svcCtx.AssetClient.SubAvailable(l.ctx, &asset.SubAvailableReq{TenantId: tenant, UserId: user, WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: coin, Amount: amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: bizID, BizNo: bizNo, Remark: remark})
	}
	if err != nil {
		return err
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("asset change rejected: %s", resp.GetBase().GetMsg())
	}
	return nil
}

// executeADL deterministically selects opposite positions by ADL rank and ID.
// It records the quantity takeover boundary; settlement at bankruptcy price is
// deliberately capped by the bankrupt position quantity.
func (l *ProcessLiquidationsLogic) executeADL(bankrupt *models.TContractPosition, contract *models.TTradeSymbolContract, liq *models.TContractLiquidation, deficit decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	positions, err := l.svcCtx.ContractPositionModel.FindList(l.ctx, models.ContractPositionPageFilter{TenantId: bankrupt.TenantId, SymbolId: bankrupt.SymbolId})
	if err != nil {
		return decimal.Zero, deficit, err
	}
	sort.SliceStable(positions, func(i, j int) bool {
		if positions[i].AdlRank == positions[j].AdlRank {
			return positions[i].Id < positions[j].Id
		}
		return positions[i].AdlRank > positions[j].AdlRank
	})
	done := decimal.Zero
	remaining := deficit
	remainingQty := decimalMaxZero(bankrupt.Qty.Sub(liq.AdlQty))
	for _, candidate := range positions {
		if candidate.Id == bankrupt.Id || !candidate.Qty.IsPositive() || candidate.PositionSide == bankrupt.PositionSide || candidate.Status != int64(trade.PositionStatus_POSITION_STATUS_NORMAL) || !remainingQty.IsPositive() {
			continue
		}
		markPnl := contractRealizedPnl(candidate.PositionSide, candidate.OpenAvgPrice, bankrupt.MarkPrice, candidate.Qty, contract.ContractSize, candidate.ContractValueType)
		bankruptcyPnl := contractRealizedPnl(candidate.PositionSide, candidate.OpenAvgPrice, bankrupt.BankruptcyPrice, candidate.Qty, contract.ContractSize, candidate.ContractValueType)
		reliefPerQty := markPnl.Sub(bankruptcyPnl).Div(candidate.Qty)
		if !reliefPerQty.IsPositive() {
			continue
		}
		qty := adlTakeoverQty(candidate.Qty, remainingQty, remaining, reliefPerQty)
		if !qty.IsPositive() {
			continue
		}
		before := cloneContractPosition(candidate)
		err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			conn := sqlx.NewSqlConnFromSession(session)
			pm := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)
			hm := models.NewTContractPositionHistoryModel(conn, l.svcCtx.Config.CacheRedis)
			current, e := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, candidate.TenantId, candidate.UserId, candidate.SymbolId, candidate.PositionSide, candidate.MarginMode)
			if e != nil {
				return e
			}
			if current.Version != before.Version || current.Qty.LessThan(qty) || current.Status != int64(trade.PositionStatus_POSITION_STATUS_NORMAL) {
				return errors.New("ADL candidate changed during takeover")
			}
			before = cloneContractPosition(current)
			positionMarginRelease, isolatedMarginRelease := adlMarginRelease(current.PositionMargin, current.IsolatedMargin, qty, current.Qty)
			pnl := contractRealizedPnl(current.PositionSide, current.OpenAvgPrice, bankrupt.BankruptcyPrice, qty, contract.ContractSize, current.ContractValueType)
			credit := decimalMaxZero(positionMarginRelease.Add(isolatedMarginRelease).Add(pnl))
			// Keep the position row locked while applying the idempotent Asset change.
			// If the local transaction fails after Asset succeeds, retrying the same
			// biz_no completes the position change without crediting twice.
			if credit.IsPositive() {
				if e = l.assetChange(current.TenantId, current.UserId, current.MarginAsset, credit, true, liq.Id, fmt.Sprintf("%s-ADL-%d", liq.LiquidationNo, current.Id), "automatic deleveraging"); e != nil {
					return e
				}
			}
			current.Qty = current.Qty.Sub(qty)
			current.AvailQty = decimalMaxZero(current.AvailQty.Sub(qty))
			current.PositionMargin = decimalMaxZero(current.PositionMargin.Sub(positionMarginRelease))
			current.IsolatedMargin = decimalMaxZero(current.IsolatedMargin.Sub(isolatedMarginRelease))
			current.RealizedPnl = current.RealizedPnl.Add(pnl)
			current.AdlRank = 0
			current.Version++
			current.UpdateTimes = utils.NowMillis()
			if current.Qty.IsZero() {
				current.Status = int64(trade.PositionStatus_POSITION_STATUS_CLOSED)
				current.ClosedAt = utils.NowMillis()
			}
			if e = pm.Update(ctx, current); e != nil {
				return e
			}
			return writeSystemPositionHistory(ctx, hm, before, current, fmt.Sprintf("%s:ADL:%d", liq.LiquidationNo, current.Id), trade.PositionActionType_POSITION_ACTION_TYPE_LIQUIDATION, pnl, decimal.Zero, bankrupt.BankruptcyPrice, "automatic deleveraging")
		})
		if err != nil {
			return done, remaining, err
		}
		done = done.Add(qty)
		remainingQty = decimalMaxZero(remainingQty.Sub(qty))
		remaining = decimalMaxZero(remaining.Sub(reliefPerQty.Mul(qty)))
		if remaining.IsZero() {
			break
		}
	}
	return done, remaining, nil
}

func adlTakeoverQty(candidateQty, bankruptRemainingQty, deficit, reliefPerQty decimal.Decimal) decimal.Decimal {
	if !candidateQty.IsPositive() || !bankruptRemainingQty.IsPositive() || !deficit.IsPositive() || !reliefPerQty.IsPositive() {
		return decimal.Zero
	}
	return decimalMin(decimalMin(candidateQty, deficit.Div(reliefPerQty)), bankruptRemainingQty)
}

func adlMarginRelease(positionMargin, isolatedMargin, qty, totalQty decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	return proportionalAmount(positionMargin, qty, totalQty), proportionalAmount(isolatedMargin, qty, totalQty)
}

func (l *ProcessLiquidationsLogic) completeLiquidation(position *models.TContractPosition, liq *models.TContractLiquidation, fee decimal.Decimal) error {
	now := utils.NowMillis()
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		pm := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)
		lm := models.NewTContractLiquidationModel(conn, l.svcCtx.Config.CacheRedis)
		em := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis)
		hm := models.NewTContractPositionHistoryModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, position.TenantId, position.UserId, position.SymbolId, position.PositionSide, position.MarginMode)
		if err != nil {
			return err
		}
		before := cloneContractPosition(current)
		liq.LiquidatedQty = current.Qty
		liq.LiquidationFee = fee
		current.Qty, current.AvailQty, current.FrozenQty = decimal.Zero, decimal.Zero, decimal.Zero
		current.PositionMargin, current.IsolatedMargin, current.MaintenanceMargin = decimal.Zero, decimal.Zero, decimal.Zero
		current.UnrealizedPnl = decimal.Zero
		current.Status = int64(trade.PositionStatus_POSITION_STATUS_CLOSED)
		current.ClosedAt = now
		current.Version++
		current.UpdateTimes = now
		if err := pm.Update(ctx, current); err != nil {
			return err
		}
		if err := writeSystemPositionHistory(ctx, hm, before, current, liq.LiquidationNo, trade.PositionActionType_POSITION_ACTION_TYPE_LIQUIDATION, liq.AccountEquity.Sub(before.PositionMargin).Sub(before.IsolatedMargin), fee, liq.TriggerMarkPrice, "forced liquidation"); err != nil {
			return err
		}
		liq.Status = int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED)
		liq.CompletedAt = now
		liq.Version++
		liq.UpdateTimes = now
		if err := lm.Update(ctx, liq); err != nil {
			return err
		}
		_, err = em.Insert(ctx, &models.TBizTradeEvent{TenantId: liq.TenantId, EventNo: liq.LiquidationNo + "-COMPLETED", EventType: "LIQUIDATION_COMPLETED", BizId: liq.LiquidationNo, BizType: "liquidation", UserId: liq.UserId, SymbolId: liq.SymbolId, ProductType: int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE), Source: int64(trade.SourceType_SOURCE_TYPE_TASK), EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20, NextRetryAt: now, Payload: "{}", CreateTimes: now, UpdateTimes: now})
		return err
	})
}
