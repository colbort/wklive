package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/domain/contractmath"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessLiquidationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProcessLiquidationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessLiquidationsLogic {
	return &ProcessLiquidationsLogic{ctx: ctx, svcCtx: svcCtx}
}

// RecoverADLExecutions resumes durable ADL asset-position sagas independently
// of the liquidation trigger path. It is safe to run concurrently: the asset
// instruction lease and the final execution row lock provide serialization.
func (l *ProcessLiquidationsLogic) RecoverADLExecutions(limit int64) error {
	rows, err := l.svcCtx.ContractAdlExecutionModel.FindRecoverable(l.ctx, limit)
	if err != nil {
		return err
	}
	var firstErr error
	for _, execution := range rows {
		if err = l.runADLExecution(execution); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// RecoverLiquidations resumes the parent saga after insurance or ADL side
// effects. Child ADL recovery alone is insufficient because it cannot close
// the bankrupt position or publish LIQUIDATION_COMPLETED.
func (l *ProcessLiquidationsLogic) RecoverLiquidations(limit int64) error {
	rows, err := l.svcCtx.ContractLiquidationModel.FindRecoverable(l.ctx, limit)
	if err != nil {
		return err
	}
	var firstErr error
	for _, liquidation := range rows {
		if err = l.ProcessPosition(liquidation.PositionId); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
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
		if !positionRequiresLiquidation(position) {
			return nil
		}
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
	if liquidation.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW) {
		return nil
	}
	if shouldHoldLiquidationForManual(l.svcCtx.Config.AutomaticLiquidation.Enabled, liquidation.Status) {
		return l.markLiquidationManual(liquidation, "automatic liquidation is disabled until P1-02 production acceptance")
	}
	lockedPosition, err := l.lockRiskUnit(position, liquidation)
	if err != nil {
		return err
	}
	if lockedPosition == nil {
		return nil
	}
	if err := l.cancelLiquidationOrders(lockedPosition); err != nil {
		return err
	}
	return l.settleTakeover(lockedPosition, contract, liquidation)
}

func positionRequiresLiquidation(position *models.TContractPosition) bool {
	if position == nil || !position.Qty.IsPositive() || !position.MarkPrice.IsPositive() || !position.LiquidationPrice.IsPositive() {
		return false
	}
	if position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_LONG) {
		return position.MarkPrice.LessThanOrEqual(position.LiquidationPrice)
	}
	return position.MarkPrice.GreaterThanOrEqual(position.LiquidationPrice)
}

func shouldHoldLiquidationForManual(enabled bool, status int64) bool {
	return !enabled && status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_PENDING_TAKEOVER)
}

func (l *ProcessLiquidationsLogic) lockRiskUnit(position *models.TContractPosition, liq *models.TContractLiquidation) (*models.TContractPosition, error) {
	var locked *models.TContractPosition
	err := l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		pm := tx.ContractPosition
		lm := tx.ContractLiquidation
		current, err := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, position.TenantId, position.UserId, position.SymbolId, position.PositionSide, position.MarginMode)
		if err != nil {
			return err
		}
		if !current.Qty.IsPositive() {
			return nil
		}
		currentLiquidation, err := lm.FindOneForUpdate(ctx, liq.Id)
		if err != nil {
			return err
		}
		if currentLiquidation.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED) || currentLiquidation.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW) {
			return nil
		}
		*liq = *currentLiquidation
		if current.Status == int64(trade.PositionStatus_POSITION_STATUS_LIQUIDATING) {
			locked = cloneContractPosition(current)
		} else {
			if current.Status != int64(trade.PositionStatus_POSITION_STATUS_NORMAL) || current.Version != position.Version {
				return errors.New("position changed before liquidation takeover")
			}
			current.Status = int64(trade.PositionStatus_POSITION_STATUS_LIQUIDATING)
			current.Version++
			current.UpdateTimes = utils.NowMillis()
			if err := pm.Update(ctx, current); err != nil {
				return err
			}
			locked = cloneContractPosition(current)
		}
		if liquidationStageRank(trade.LiquidationStatus(liq.Status)) < liquidationStageRank(trade.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING) {
			liq.Status = int64(trade.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING)
		}
		if liq.StartedAt == 0 {
			liq.StartedAt = utils.NowMillis()
		}
		liq.Version++
		liq.UpdateTimes = utils.NowMillis()
		return lm.Update(ctx, liq)
	})
	return locked, err
}

// A full takeover must remove every active order for the risk unit. In
// particular, reduce-only orders are no longer executable after the position
// is closed and their reserved close quantity must be released first.
func (l *ProcessLiquidationsLogic) cancelLiquidationOrders(position *models.TContractPosition) error {
	cursor := int64(0)
	for {
		statuses := append(helpers.MatchableOrderStatuses(), int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING))
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{TenantId: position.TenantId, UserId: position.UserId, SymbolId: position.SymbolId, ProductType: int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE), Statuses: statuses}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, o := range orders {
			cursor = o.Id
			terminating, err := beginSystemOrderTermination(l.ctx, l.svcCtx, o.Id, "risk liquidation", false)
			if err != nil {
				return err
			}
			if terminating == nil {
				continue
			}
			if err := unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, terminating, "liquidation risk order release"); err != nil {
				return err
			}
			if err := removeOrderBookOrder(l.svcCtx, l.ctx, terminating); err != nil {
				logx.WithContext(l.ctx).Errorf("remove liquidation order from cache failed, orderId=%d err=%v", terminating.Id, err)
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessLiquidationsLogic) settleTakeover(position *models.TContractPosition, contract *models.TTradeSymbolContract, liq *models.TContractLiquidation) error {
	partial, err := l.tryPartialTakeover(position, contract, liq)
	if err != nil {
		return err
	}
	if partial {
		return nil
	}
	values, err := contractmath.CalculateTradeValues(position.ContractValueType, liq.TriggerQty, contract.ContractSize, liq.TriggerMarkPrice)
	if err != nil {
		return err
	}
	nominalFee := contractmath.RoundDebit(values.SettlementNotional.Mul(contract.LiquidationFeeRate))
	fee, residual, deficit := splitLiquidationEquity(liq.AccountEquity, nominalFee)
	if fee.IsPositive() {
		if err := l.creditPlatformRevenue(position.TenantId, position.MarginAsset, fee, liq); err != nil {
			return err
		}
	}
	if residual.IsPositive() {
		if err := l.assetChange(position.TenantId, position.UserId, position.MarginAsset, residual, true, liq.Id, liq.LiquidationNo+"-RESIDUAL", "liquidation residual equity"); err != nil {
			return err
		}
	}
	var fund *models.TContractInsuranceFundAccount
	if deficit.IsPositive() {
		fund, err = l.svcCtx.ContractInsuranceFundModel.FindEnabled(l.ctx, position.TenantId, position.SymbolId, position.MarginAsset)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if err == nil {
			if err = l.checkpointLiquidation(liq, trade.LiquidationStatus_LIQUIDATION_STATUS_INSURANCE_FUND, liq.InsuranceFundAmount, liq.AdlQty); err != nil {
				return err
			}
			covered, remaining, coverErr := l.tryInsurance(fund, deficit, liq)
			if coverErr != nil {
				return coverErr
			}
			liq.InsuranceFundAmount = covered
			if err = l.checkpointLiquidation(liq, trade.LiquidationStatus_LIQUIDATION_STATUS_INSURANCE_FUND, covered, liq.AdlQty); err != nil {
				return err
			}
			deficit = remaining
		}
	}
	if deficit.IsPositive() {
		if fund == nil || fund.AdlEnabled != int64(common.YesNo_YES_NO_YES) {
			return l.markLiquidationManual(liq, "insurance fund unavailable or insufficient")
		}
		if err = l.checkpointLiquidation(liq, trade.LiquidationStatus_LIQUIDATION_STATUS_ADL, liq.InsuranceFundAmount, liq.AdlQty); err != nil {
			return err
		}
		adlQty, remaining, err := l.executeADL(position, contract, liq, deficit)
		if err != nil {
			return err
		}
		liq.AdlQty = adlQty
		if err = l.checkpointLiquidation(liq, trade.LiquidationStatus_LIQUIDATION_STATUS_ADL, liq.InsuranceFundAmount, adlQty); err != nil {
			return err
		}
		if remaining.IsPositive() {
			return l.markLiquidationManual(liq, "ADL liquidity insufficient")
		}
	}
	return l.completeLiquidation(position, liq, fee)
}

type partialLiquidationPlan struct {
	after            *models.TContractPosition
	liquidatedQty    decimal.Decimal
	marginRelease    decimal.Decimal
	isolatedRelease  decimal.Decimal
	realizedPnl      decimal.Decimal
	fee              decimal.Decimal
	availableCredit  decimal.Decimal
	targetTierNumber int64
}

func (l *ProcessLiquidationsLogic) tryPartialTakeover(position *models.TContractPosition, contract *models.TTradeSymbolContract, liq *models.TContractLiquidation) (bool, error) {
	values, err := contractmath.CalculateTradeValues(position.ContractValueType, liq.TriggerQty, contract.ContractSize, liq.TriggerMarkPrice)
	if err != nil {
		return false, err
	}
	currentTier, err := l.svcCtx.ContractRiskLimitTierModel.FindByNotional(l.ctx, position.TenantId, position.SymbolId, values.QuoteNotional)
	if errors.Is(err, models.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	tiers, _, err := l.svcCtx.ContractRiskLimitTierModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: position.TenantId, SymbolId: position.SymbolId, Enabled: 1}, 0, 100)
	if err != nil {
		return false, err
	}
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, position.SymbolId)
	if err != nil {
		return false, err
	}
	plan := buildPartialLiquidationPlan(position, contract, symbol, currentTier, tiers)
	if plan == nil {
		return false, nil
	}
	if plan.fee.IsPositive() {
		if err = l.creditPlatformRevenueByBizNo(position.TenantId, position.MarginAsset, plan.fee, liq.Id, liq.LiquidationNo+"-PARTIAL-FEE", "partial liquidation fee revenue"); err != nil {
			return false, err
		}
	}
	if plan.availableCredit.IsPositive() {
		if err = l.assetChange(position.TenantId, position.UserId, position.MarginAsset, plan.availableCredit, true, liq.Id, liq.LiquidationNo+"-PARTIAL-RESIDUAL", "partial liquidation released equity"); err != nil {
			return false, err
		}
	}
	if err = l.completePartialLiquidation(position, liq, plan); err != nil {
		return false, err
	}
	return true, nil
}

func buildPartialLiquidationPlan(position *models.TContractPosition, contract *models.TTradeSymbolContract, symbol *models.TTradeSymbol, currentTier *models.TContractRiskLimitTier, tiers []*models.TContractRiskLimitTier) *partialLiquidationPlan {
	if position == nil || contract == nil || symbol == nil || currentTier == nil || currentTier.TierNo <= 1 || !symbol.QtyStep.IsPositive() {
		return nil
	}
	sort.SliceStable(tiers, func(i, j int) bool { return tiers[i].TierNo > tiers[j].TierNo })
	for _, tier := range tiers {
		if tier == nil || tier.Enabled != 1 || tier.TierNo >= currentTier.TierNo || !tier.NotionalCap.IsPositive() {
			continue
		}
		var targetQty decimal.Decimal
		if position.ContractValueType == int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE) {
			targetQty = tier.NotionalCap.Div(contract.ContractSize)
		} else {
			targetQty = tier.NotionalCap.Div(contract.ContractSize.Mul(position.MarkPrice))
		}
		targetQty = targetQty.Div(symbol.QtyStep).Floor().Mul(symbol.QtyStep)
		if !targetQty.IsPositive() || !targetQty.LessThan(position.Qty) {
			continue
		}
		liquidatedQty := position.Qty.Sub(targetQty)
		marginRelease, isolatedRelease := adlMarginRelease(position.PositionMargin, position.IsolatedMargin, liquidatedQty, position.Qty)
		realized := contractRealizedPnl(position.PositionSide, position.OpenAvgPrice, position.MarkPrice, liquidatedQty, contract.ContractSize, position.ContractValueType)
		closedValues, err := contractmath.CalculateTradeValues(position.ContractValueType, liquidatedQty, contract.ContractSize, position.MarkPrice)
		if err != nil {
			continue
		}
		fee := contractmath.RoundDebit(closedValues.SettlementNotional.Mul(contract.LiquidationFeeRate))
		releasedEquity := marginRelease.Add(isolatedRelease).Add(realized)
		if releasedEquity.LessThan(fee) {
			// A partial close that consumes margin belonging to the remaining
			// position cannot safely restore the risk unit. Fall back to full
			// takeover where insurance/ADL are available.
			continue
		}
		after := cloneContractPosition(position)
		after.Qty = targetQty
		after.AvailQty = decimalMin(after.AvailQty, targetQty)
		after.FrozenQty = decimal.Zero
		after.PositionMargin = decimalMaxZero(after.PositionMargin.Sub(marginRelease))
		after.IsolatedMargin = decimalMaxZero(after.IsolatedMargin.Sub(isolatedRelease))
		after.RealizedPnl = after.RealizedPnl.Add(realized)
		recalculatePositionRisk(after, contract, tier)
		if !positionRiskRecovered(after) {
			continue
		}
		return &partialLiquidationPlan{
			after:            after,
			liquidatedQty:    liquidatedQty,
			marginRelease:    marginRelease,
			isolatedRelease:  isolatedRelease,
			realizedPnl:      realized,
			fee:              fee,
			availableCredit:  releasedEquity.Sub(fee),
			targetTierNumber: tier.TierNo,
		}
	}
	return nil
}

func positionRiskRecovered(position *models.TContractPosition) bool {
	if position == nil || !position.Qty.IsPositive() {
		return false
	}
	equity := position.PositionMargin.Add(position.IsolatedMargin).Add(position.UnrealizedPnl)
	if !equity.GreaterThan(position.MaintenanceMargin) || !position.RiskRate.LessThan(decimal.NewFromInt(1)) {
		return false
	}
	if position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_LONG) {
		return position.MarkPrice.GreaterThan(position.LiquidationPrice)
	}
	return position.MarkPrice.LessThan(position.LiquidationPrice)
}

func (l *ProcessLiquidationsLogic) completePartialLiquidation(position *models.TContractPosition, liq *models.TContractLiquidation, plan *partialLiquidationPlan) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		pm := tx.ContractPosition
		lm := tx.ContractLiquidation
		hm := tx.ContractPositionHistory
		em := tx.TradeEventOutbox
		current, err := pm.FindOneForUpdate(ctx, position.Id)
		if err != nil {
			return err
		}
		currentLiquidation, err := lm.FindOneForUpdate(ctx, liq.Id)
		if err != nil {
			return err
		}
		if currentLiquidation.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_PARTIAL_RECOVERED) {
			return nil
		}
		if current.Status != int64(trade.PositionStatus_POSITION_STATUS_LIQUIDATING) ||
			current.Version != position.Version || !current.Qty.Equal(liq.TriggerQty) {
			return errors.New("position changed before partial liquidation commit")
		}
		before := cloneContractPosition(current)
		after := cloneContractPosition(plan.after)
		after.Status = int64(trade.PositionStatus_POSITION_STATUS_NORMAL)
		after.Version = current.Version + 1
		after.UpdateTimes = now
		if err = pm.Update(ctx, after); err != nil {
			return err
		}
		if err = writeSystemPositionHistory(ctx, hm, before, after, liq.CreateTimes, liq.LiquidationNo, trade.PositionActionType_POSITION_ACTION_TYPE_LIQUIDATION, plan.realizedPnl, plan.fee, liq.TriggerMarkPrice, "partial liquidation recovered risk tier"); err != nil {
			return err
		}
		currentLiquidation.LiquidatedQty = plan.liquidatedQty
		currentLiquidation.LiquidationFee = plan.fee
		currentLiquidation.Status = int64(trade.LiquidationStatus_LIQUIDATION_STATUS_PARTIAL_RECOVERED)
		currentLiquidation.Reason = fmt.Sprintf("partial liquidation restored risk at tier %d", plan.targetTierNumber)
		currentLiquidation.CompletedAt = now
		currentLiquidation.Version++
		currentLiquidation.UpdateTimes = now
		if err = lm.Update(ctx, currentLiquidation); err != nil {
			return err
		}
		_, err = em.Insert(ctx, &models.TTradeEventOutbox{TenantId: liq.TenantId, EventNo: liq.LiquidationNo + "-COMPLETED", EventType: "LIQUIDATION_PARTIAL_RECOVERED", BizId: liq.LiquidationNo, BizType: "liquidation", UserId: liq.UserId, SymbolId: liq.SymbolId, ProductType: int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE), Source: int64(trade.SourceType_SOURCE_TYPE_TASK), EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20, NextRetryAt: now, Payload: "{}", CreateTimes: now, UpdateTimes: now})
		return err
	})
}

// splitLiquidationEquity ensures that a liquidation fee can only be collected
// from positive remaining equity. Insurance and ADL cover bankruptcy loss, not
// platform fee revenue.
func splitLiquidationEquity(grossEquity, nominalFee decimal.Decimal) (fee, residual, deficit decimal.Decimal) {
	positiveEquity := decimalMaxZero(grossEquity)
	fee = decimalMin(decimalMaxZero(nominalFee), positiveEquity)
	residual = decimalMaxZero(positiveEquity.Sub(fee))
	deficit = decimalMaxZero(grossEquity.Neg())
	return fee, residual, deficit
}

func (l *ProcessLiquidationsLogic) markLiquidationManual(liq *models.TContractLiquidation, reason string) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		lm := tx.ContractLiquidation
		current, err := lm.FindOneForUpdate(ctx, liq.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED) {
			*liq = *current
			return nil
		}
		if liq.InsuranceFundAmount.GreaterThan(current.InsuranceFundAmount) {
			current.InsuranceFundAmount = liq.InsuranceFundAmount
		}
		if liq.AdlQty.GreaterThan(current.AdlQty) {
			current.AdlQty = liq.AdlQty
		}
		current.Status = int64(trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW)
		current.Reason = reason
		current.Version++
		current.UpdateTimes = now
		if err = lm.Update(ctx, current); err != nil {
			return err
		}
		*liq = *current
		return nil
	})
}

func (l *ProcessLiquidationsLogic) checkpointLiquidation(liq *models.TContractLiquidation, status trade.LiquidationStatus, insuranceAmount, adlQty decimal.Decimal) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		lm := tx.ContractLiquidation
		current, err := lm.FindOneForUpdate(ctx, liq.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED) || current.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW) {
			*liq = *current
			return nil
		}
		if liquidationStageRank(status) > liquidationStageRank(trade.LiquidationStatus(current.Status)) {
			current.Status = int64(status)
		}
		if insuranceAmount.GreaterThan(current.InsuranceFundAmount) {
			current.InsuranceFundAmount = insuranceAmount
		}
		if adlQty.GreaterThan(current.AdlQty) {
			current.AdlQty = adlQty
		}
		current.Version++
		current.UpdateTimes = now
		if err = lm.Update(ctx, current); err != nil {
			return err
		}
		*liq = *current
		return nil
	})
}

func liquidationStageRank(status trade.LiquidationStatus) int {
	switch status {
	case trade.LiquidationStatus_LIQUIDATION_STATUS_PENDING_TAKEOVER:
		return 1
	case trade.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING,
		trade.LiquidationStatus_LIQUIDATION_STATUS_PARTIAL_RECOVERED:
		return 2
	case trade.LiquidationStatus_LIQUIDATION_STATUS_INSURANCE_FUND:
		return 3
	case trade.LiquidationStatus_LIQUIDATION_STATUS_ADL:
		return 4
	case trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED:
		return 5
	case trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW:
		return 6
	default:
		return 0
	}
}

func (l *ProcessLiquidationsLogic) tryInsurance(fund *models.TContractInsuranceFundAccount, amount decimal.Decimal, liq *models.TContractLiquidation) (decimal.Decimal, decimal.Decimal, error) {
	resp, err := l.svcCtx.AssetClient.CoverInsuranceDeficit(l.ctx, &asset.CoverInsuranceDeficitReq{TenantId: fund.TenantId, Coin: fund.SettleAsset, RequestedAmount: amount.String(), LiquidationId: liq.Id, LiquidationNo: liq.LiquidationNo + "-INSURANCE", Remark: "liquidation insurance fund cover"})
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
	return validateLiquidationAssetResponse(resp)
}

func (l *ProcessLiquidationsLogic) creditPlatformRevenue(tenant int64, coin string, amount decimal.Decimal, liq *models.TContractLiquidation) error {
	return l.creditPlatformRevenueByBizNo(tenant, coin, amount, liq.Id, liq.LiquidationNo+"-FEE", "liquidation fee revenue")
}

func (l *ProcessLiquidationsLogic) creditPlatformRevenueByBizNo(tenant int64, coin string, amount decimal.Decimal, bizID int64, bizNo, remark string) error {
	resp, err := l.svcCtx.AssetClient.CreditPlatformRevenue(l.ctx, &asset.CreditPlatformRevenueReq{
		TenantId:  tenant,
		Coin:      coin,
		Amount:    amount.String(),
		BizType:   asset.BizType_BIZ_TYPE_TRADE,
		SceneType: asset.SceneType_SCENE_TYPE_TRADE_FEE,
		BizId:     bizID,
		BizNo:     bizNo,
		Remark:    remark,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.GetBase() == nil {
		return errors.New("platform revenue returned an empty response")
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("platform revenue rejected: %s", resp.GetBase().GetMsg())
	}
	return nil
}

func validateLiquidationAssetResponse(resp *asset.ChangeAssetResp) error {
	if resp == nil || resp.GetBase() == nil {
		return errors.New("asset change returned an empty response")
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
	existing, err := l.svcCtx.ContractAdlExecutionModel.FindByLiquidation(l.ctx, bankrupt.TenantId, liq.Id)
	if err != nil {
		return decimal.Zero, deficit, err
	}
	done, relieved := decimal.Zero, decimal.Zero
	used := make(map[int64]struct{}, len(existing))
	for _, execution := range existing {
		used[execution.PositionId] = struct{}{}
		if execution.Status != 3 {
			if err = l.runADLExecution(execution); err != nil {
				return done, decimalMaxZero(deficit.Sub(relieved)), err
			}
		}
		done, relieved = done.Add(execution.Qty), relieved.Add(execution.ReliefAmount)
	}
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
	remaining := decimalMaxZero(deficit.Sub(relieved))
	recordedQty := liq.AdlQty
	if done.GreaterThan(recordedQty) {
		recordedQty = done
	}
	remainingQty := decimalMaxZero(bankrupt.Qty.Sub(recordedQty))
	for _, candidate := range positions {
		_, alreadyUsed := used[candidate.Id]
		if alreadyUsed || candidate.Id == bankrupt.Id || !candidate.Qty.IsPositive() ||
			candidate.PositionSide == bankrupt.PositionSide ||
			candidate.MarginMode != int64(trade.MarginMode_MARGIN_MODE_ISOLATED) ||
			candidate.Status != int64(trade.PositionStatus_POSITION_STATUS_NORMAL) ||
			!remainingQty.IsPositive() {
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
		execution, prepareErr := l.prepareADLExecution(candidate, bankrupt, contract, liq, qty, reliefPerQty.Mul(qty))
		if prepareErr != nil {
			return done, remaining, prepareErr
		}
		err = l.runADLExecution(execution)
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

func (l *ProcessLiquidationsLogic) prepareADLExecution(candidate, bankrupt *models.TContractPosition, contract *models.TTradeSymbolContract, liq *models.TContractLiquidation, qty, relief decimal.Decimal) (*models.TContractAdlExecution, error) {
	now := utils.NowMillis()
	executionNo := fmt.Sprintf("%s-ADL-%d", liq.LiquidationNo, candidate.Id)
	var prepared *models.TContractAdlExecution
	err := l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		pm := tx.ContractPosition
		current, err := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, candidate.TenantId, candidate.UserId, candidate.SymbolId, candidate.PositionSide, candidate.MarginMode)
		if err != nil {
			return err
		}
		if current.Status != int64(trade.PositionStatus_POSITION_STATUS_NORMAL) || current.Version != candidate.Version || current.Qty.LessThan(qty) {
			return errors.New("ADL candidate changed before reservation")
		}
		positionMargin, isolatedMargin := adlMarginRelease(current.PositionMargin, current.IsolatedMargin, qty, current.Qty)
		pnl := contractRealizedPnl(current.PositionSide, current.OpenAvgPrice, bankrupt.BankruptcyPrice, qty, contract.ContractSize, current.ContractValueType)
		credit := decimalMaxZero(positionMargin.Add(isolatedMargin).Add(pnl))
		current.Status = int64(trade.PositionStatus_POSITION_STATUS_LIQUIDATING)
		current.Version++
		current.UpdateTimes = now
		if err = pm.Update(ctx, current); err != nil {
			return err
		}
		prepared = &models.TContractAdlExecution{TenantId: current.TenantId, ExecutionNo: executionNo, LiquidationId: liq.Id, LiquidationNo: liq.LiquidationNo, PositionId: current.Id, UserId: current.UserId, PositionVersion: current.Version, Qty: qty, PositionMarginRelease: positionMargin, IsolatedMarginRelease: isolatedMargin, RealizedPnl: pnl, AssetCredit: credit, Asset: current.MarginAsset, BankruptcyPrice: bankrupt.BankruptcyPrice, ReliefAmount: relief, Status: 1, CreateTimes: now, UpdateTimes: now}
		result, insertErr := tx.ContractAdlExecution.Insert(ctx, prepared)
		if insertErr != nil {
			return insertErr
		}
		prepared.Id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		if credit.IsPositive() {
			return insertSettlementInstructionIdempotent(ctx, tx.TradeSettlementInstruction, &models.TTradeSettlementInstruction{TenantId: current.TenantId, InstructionNo: executionNo, BizType: "adl", BizId: executionNo, BatchNo: liq.LiquidationNo, PositionId: current.Id, UserId: current.UserId, Action: int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE), Asset: current.MarginAsset, Amount: credit, StepNo: 1, Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now})
		}
		return nil
	})
	return prepared, err
}

func (l *ProcessLiquidationsLogic) runADLExecution(execution *models.TContractAdlExecution) error {
	if execution.Status == 3 {
		return nil
	}
	if execution.AssetCredit.IsPositive() {
		instruction, err := l.svcCtx.TradeSettlementInstrModel.FindOneByTenantIdInstructionNo(l.ctx, execution.TenantId, execution.ExecutionNo)
		if err != nil {
			return err
		}
		if execution.Status == 2 {
			if instruction.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
				return errors.New("ADL asset checkpoint has no successful instruction")
			}
			return l.completeADLExecution(execution)
		}
		claimed, lease, err := l.svcCtx.TradeSettlementInstrModel.ClaimLease(l.ctx, instruction.Id, utils.NowMillis())
		if err != nil {
			return err
		}
		if !claimed {
			if instruction.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
				if err = l.checkpointADLAssetDone(execution, instruction); err != nil {
					return err
				}
				return l.completeADLExecution(execution)
			}
			return fmt.Errorf("ADL instruction is not claimable: status=%d", instruction.Status)
		}
		instruction.UpdateTimes = lease
		if err = executeSimpleAssetInstruction(l.ctx, l.svcCtx, instruction, "automatic deleveraging"); err != nil {
			cause := err
			if markErr := failContractSagaInstruction(l.ctx, l.svcCtx, instruction, cause, func(ctx context.Context, tx *models.TransactionModels, current *models.TTradeSettlementInstruction, manual bool, now int64) error {
				em := tx.ContractAdlExecution
				e, findErr := em.FindOneByExecutionNo(ctx, current.TenantId, current.BizId)
				if findErr != nil {
					return findErr
				}
				e.Status, e.LastErrorMsg, e.UpdateTimes = 4, current.LastErrorMsg, now
				if manual {
					e.Status = 5
				}
				return em.Update(ctx, e)
			}); markErr != nil {
				return markErr
			}
			return cause
		}
		if err = l.checkpointADLAssetDone(execution, instruction); err != nil {
			return err
		}
	}
	return l.completeADLExecution(execution)
}

// checkpointADLAssetDone durably separates the external Asset side effect from
// the local Position projection. A crash after this commit resumes from status
// ASSET_DONE without issuing another credit.
func (l *ProcessLiquidationsLogic) checkpointADLAssetDone(execution *models.TContractAdlExecution, instruction *models.TTradeSettlementInstruction) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		em := tx.ContractAdlExecution
		e, err := em.FindOneForUpdate(ctx, execution.Id)
		if err != nil {
			return err
		}
		if e.Status == 3 {
			return nil
		}
		im := tx.TradeSettlementInstruction
		i, err := im.FindOneForUpdate(ctx, instruction.Id)
		if err != nil {
			return err
		}
		if i.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			if !settlementInstructionLeaseOwned(i, instruction) {
				return errors.New("ADL instruction lease lost")
			}
			i.Status, i.NextRetryAt, i.LastErrorMsg, i.UpdateTimes = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS), 0, "", now
			if err = im.Update(ctx, i); err != nil {
				return err
			}
		}
		e.Status, e.LastErrorMsg, e.UpdateTimes = 2, "", now
		return em.Update(ctx, e)
	})
}

func (l *ProcessLiquidationsLogic) completeADLExecution(execution *models.TContractAdlExecution) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		em := tx.ContractAdlExecution
		e, err := em.FindOneForUpdate(ctx, execution.Id)
		if err != nil {
			return err
		}
		if e.Status == 3 {
			return nil
		}
		if e.AssetCredit.IsPositive() {
			if e.Status != 2 {
				return errors.New("ADL position projection requires ASSET_DONE checkpoint")
			}
			im := tx.TradeSettlementInstruction
			instruction, findErr := im.FindOneByTenantIdInstructionNo(ctx, e.TenantId, e.ExecutionNo)
			if findErr != nil {
				return findErr
			}
			currentInstruction, lockErr := im.FindOneForUpdate(ctx, instruction.Id)
			if lockErr != nil {
				return lockErr
			}
			if currentInstruction.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
				return errors.New("ADL position projection requires successful Asset instruction")
			}
		}
		pm := tx.ContractPosition
		current, err := pm.FindOneForUpdate(ctx, e.PositionId)
		if err != nil {
			return err
		}
		if current.Status != int64(trade.PositionStatus_POSITION_STATUS_LIQUIDATING) || current.Version != e.PositionVersion || current.Qty.LessThan(e.Qty) {
			return errors.New("ADL reserved position changed")
		}
		before := cloneContractPosition(current)
		current.Qty = current.Qty.Sub(e.Qty)
		current.AvailQty = decimalMaxZero(current.AvailQty.Sub(e.Qty))
		current.PositionMargin = decimalMaxZero(current.PositionMargin.Sub(e.PositionMarginRelease))
		current.IsolatedMargin = decimalMaxZero(current.IsolatedMargin.Sub(e.IsolatedMarginRelease))
		current.RealizedPnl = current.RealizedPnl.Add(e.RealizedPnl)
		if current.Qty.IsPositive() {
			contract, err := tx.TradeSymbolContract.
				FindOneByTenantIdSymbolId(ctx, current.TenantId, current.SymbolId)
			if err != nil {
				return err
			}
			values, err := contractmath.CalculateTradeValues(current.ContractValueType, current.Qty, contract.ContractSize, current.MarkPrice)
			if err != nil {
				return err
			}
			tier, tierErr := tx.ContractRiskLimitTier.
				FindByNotional(ctx, current.TenantId, current.SymbolId, values.QuoteNotional)
			if tierErr != nil && !errors.Is(tierErr, models.ErrNotFound) {
				return tierErr
			}
			if errors.Is(tierErr, models.ErrNotFound) {
				tier = nil
			}
			recalculatePositionRisk(current, contract, tier)
		} else {
			current.MaintenanceMargin = decimal.Zero
			current.UnrealizedPnl = decimal.Zero
			current.LiquidationPrice = decimal.Zero
			current.BankruptcyPrice = decimal.Zero
			current.RiskRate = decimal.Zero
			current.AdlRank = 0
		}
		current.Version, current.UpdateTimes = current.Version+1, now
		current.Status = int64(trade.PositionStatus_POSITION_STATUS_NORMAL)
		if current.Qty.IsZero() {
			current.Status, current.ClosedAt = int64(trade.PositionStatus_POSITION_STATUS_CLOSED), now
		}
		if err = pm.Update(ctx, current); err != nil {
			return err
		}
		if err = writeSystemPositionHistory(ctx, tx.ContractPositionHistory, before, current, e.CreateTimes, e.ExecutionNo, trade.PositionActionType_POSITION_ACTION_TYPE_LIQUIDATION, e.RealizedPnl, decimal.Zero, e.BankruptcyPrice, "automatic deleveraging"); err != nil {
			return err
		}
		e.Status, e.LastErrorMsg, e.UpdateTimes = 3, "", now
		return em.Update(ctx, e)
	})
}

func (l *ProcessLiquidationsLogic) completeLiquidation(position *models.TContractPosition, liq *models.TContractLiquidation, fee decimal.Decimal) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		pm := tx.ContractPosition
		lm := tx.ContractLiquidation
		em := tx.TradeEventOutbox
		hm := tx.ContractPositionHistory
		current, err := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, position.TenantId, position.UserId, position.SymbolId, position.PositionSide, position.MarginMode)
		if err != nil {
			return err
		}
		currentLiquidation, err := lm.FindOneForUpdate(ctx, liq.Id)
		if err != nil {
			return err
		}
		if currentLiquidation.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED) {
			return nil
		}
		*liq = *currentLiquidation
		before := cloneContractPosition(current)
		liq.LiquidatedQty = current.Qty
		liq.LiquidationFee = fee
		current.Qty, current.AvailQty, current.FrozenQty = decimal.Zero, decimal.Zero, decimal.Zero
		current.PositionMargin, current.IsolatedMargin, current.MaintenanceMargin = decimal.Zero, decimal.Zero, decimal.Zero
		current.UnrealizedPnl = decimal.Zero
		current.LiquidationPrice, current.BankruptcyPrice, current.RiskRate = decimal.Zero, decimal.Zero, decimal.Zero
		current.AdlRank = 0
		current.Status = int64(trade.PositionStatus_POSITION_STATUS_CLOSED)
		current.ClosedAt = now
		current.Version++
		current.UpdateTimes = now
		if err := pm.Update(ctx, current); err != nil {
			return err
		}
		if err := writeSystemPositionHistory(ctx, hm, before, current, liq.CreateTimes, liq.LiquidationNo, trade.PositionActionType_POSITION_ACTION_TYPE_LIQUIDATION, liq.AccountEquity.Sub(before.PositionMargin).Sub(before.IsolatedMargin), fee, liq.TriggerMarkPrice, "forced liquidation"); err != nil {
			return err
		}
		liq.Status = int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED)
		liq.CompletedAt = now
		liq.Version++
		liq.UpdateTimes = now
		if err := lm.Update(ctx, liq); err != nil {
			return err
		}
		_, err = em.Insert(ctx, &models.TTradeEventOutbox{TenantId: liq.TenantId, EventNo: liq.LiquidationNo + "-COMPLETED", EventType: "LIQUIDATION_COMPLETED", BizId: liq.LiquidationNo, BizType: "liquidation", UserId: liq.UserId, SymbolId: liq.SymbolId, ProductType: int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE), Source: int64(trade.SourceType_SOURCE_TYPE_TASK), EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20, NextRetryAt: now, Payload: "{}", CreateTimes: now, UpdateTimes: now})
		return err
	})
}
