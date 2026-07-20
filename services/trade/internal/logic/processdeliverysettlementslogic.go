package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessDeliverySettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProcessDeliverySettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessDeliverySettlementsLogic {
	return &ProcessDeliverySettlementsLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ProcessDeliverySettlementsLogic) Process(tenantID int64) error {
	if err := l.advanceSymbols(tenantID); err != nil {
		return err
	}
	return l.settlePending(tenantID)
}

func (l *ProcessDeliverySettlementsLogic) advanceSymbols(tenantID int64) error {
	now := utils.NowMillis()
	cursor := int64(0)
	for {
		contracts, _, err := l.svcCtx.TradeSymbolContractModel.FindPage(l.ctx, cursor, 100)
		if err != nil {
			return err
		}
		if len(contracts) == 0 {
			return nil
		}
		for _, c := range contracts {
			cursor = c.Id
			if tenantID > 0 && c.TenantId != tenantID || c.DeliveryTime <= 0 {
				continue
			}
			symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, c.SymbolId)
			if err != nil {
				return err
			}
			if symbol.ContractType != int64(trade.ContractType_CONTRACT_TYPE_DELIVERY) {
				continue
			}
			if c.OpenCutoffTime > 0 && now >= c.OpenCutoffTime && symbol.Status == int64(trade.SymbolStatus_SYMBOL_STATUS_ENABLED) {
				symbol.Status = int64(trade.SymbolStatus_SYMBOL_STATUS_CLOSE_ONLY)
				symbol.UpdateTimes = now
				if err := l.svcCtx.TradeSymbolModel.Update(l.ctx, symbol); err != nil {
					return err
				}
			}
			if c.MatchingStopTime > 0 && now >= c.MatchingStopTime {
				if err := l.cancelSymbolOrders(symbol); err != nil {
					return err
				}
			}
			if now < c.DeliveryTime {
				continue
			}
			unfinished, err := l.svcCtx.TradeOrderModel.CountBySymbolStatuses(l.ctx, symbol.TenantId, symbol.Id, []int64{int64(trade.OrderStatus_ORDER_STATUS_PENDING), int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED), int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING), int64(trade.OrderStatus_ORDER_STATUS_CANCELING), int64(trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING)})
			if err != nil {
				return err
			}
			if unfinished > 0 {
				continue
			}
			if err := l.ensureBatch(symbol, c); err != nil {
				return err
			}
		}
		if len(contracts) < 100 {
			return nil
		}
	}
}

func (l *ProcessDeliverySettlementsLogic) cancelSymbolOrders(symbol *models.TTradeSymbol) error {
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{TenantId: symbol.TenantId, SymbolId: symbol.Id, ProductType: int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE), Statuses: []int64{int64(trade.OrderStatus_ORDER_STATUS_PENDING), int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED), int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING)}}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, o := range orders {
			cursor = o.Id
			now := utils.NowMillis()
			o.Status = int64(trade.OrderStatus_ORDER_STATUS_CANCELING)
			o.CanceledQty = decimalMaxZero(o.Qty.Sub(o.FilledQty))
			o.CancelReason = "delivery matching stopped"
			o.Version++
			o.UpdateTimes = now
			if err := l.svcCtx.TradeOrderModel.Update(l.ctx, o); err != nil {
				return err
			}
			if err := removeOrderBookOrder(l.svcCtx, l.ctx, o); err != nil {
				return err
			}
			if err := unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, o, "delivery order release"); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessDeliverySettlementsLogic) ensureBatch(symbol *models.TTradeSymbol, c *models.TTradeSymbolContract) error {
	if _, err := l.svcCtx.ContractDeliveryBatchModel.FindOneByTenantIdSymbolIdDeliveryTime(l.ctx, c.TenantId, c.SymbolId, c.DeliveryTime); err == nil {
		return nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return err
	}
	quote, _, err := NewProcessSecondsSettlementsLogic(l.ctx, l.svcCtx).getValidQuotesAtKind("DELIVERY_PRICE", c.SettlementPriceSource, c.SymbolId, c.DeliveryTime, c.SettlementWindowSeconds*1000)
	if err != nil {
		return err
	}
	window := c.SettlementWindowSeconds * 1000
	if window <= 0 || quote.QuoteTs < c.DeliveryTime-window || quote.QuoteTs > c.DeliveryTime+window {
		return fmt.Errorf("delivery quote outside configured settlement window")
	}
	price := mustParseFloat(quote.LastPrice)
	positions, err := l.svcCtx.ContractPositionModel.FindList(l.ctx, models.ContractPositionPageFilter{TenantId: c.TenantId, SymbolId: c.SymbolId, ContractType: int64(trade.ContractType_CONTRACT_TYPE_DELIVERY)})
	if err != nil {
		return err
	}
	active := positions[:0]
	for _, p := range positions {
		if p.Qty.IsPositive() {
			active = append(active, p)
		}
	}
	now := utils.NowMillis()
	batchNo := fmt.Sprintf("DLV-%d-%d", c.SymbolId, c.DeliveryTime)
	raw, _ := normalizeJSON(map[string]any{"quote": quote})
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		bm := models.NewTContractDeliveryBatchModel(conn, l.svcCtx.Config.CacheRedis)
		sm := models.NewTContractDeliverySettlementModel(conn, l.svcCtx.Config.CacheRedis)
		pm := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)
		symbolModel := models.NewTTradeSymbolModel(conn, l.svcCtx.Config.CacheRedis)
		res, err := bm.Insert(ctx, &models.TContractDeliveryBatch{TenantId: c.TenantId, BatchNo: batchNo, SymbolId: c.SymbolId, SettlementPrice: price, PriceSource: quote.SnapshotID, PriceAlgorithm: nonEmpty(c.SettlementPriceAlgorithm, "last-v1"), SampleSnapshot: sql.NullString{String: raw, Valid: true}, OpenCutoffTime: c.OpenCutoffTime, MatchingStopTime: c.MatchingStopTime, DeliveryTime: c.DeliveryTime, Status: int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING), TotalPositions: int64(len(active)), CreateTimes: now, UpdateTimes: now})
		if err != nil {
			return err
		}
		batchID, _ := res.LastInsertId()
		for _, p := range active {
			locked, err := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, p.TenantId, p.UserId, p.SymbolId, p.PositionSide, p.MarginMode)
			if err != nil {
				return err
			}
			if !locked.Qty.Equal(p.Qty) {
				return fmt.Errorf("delivery position changed during price lock: %d", p.Id)
			}
			pnl := contractRealizedPnl(locked.PositionSide, locked.OpenAvgPrice, price, locked.Qty, c.ContractSize, locked.ContractValueType)
			values, err := calculateContractTradeValues(locked.ContractValueType, locked.Qty, c.ContractSize, price)
			if err != nil {
				return err
			}
			fee := roundContractDebit(values.SettlementNotional.Mul(c.DeliveryFeeRate))
			if _, err = sm.Insert(ctx, &models.TContractDeliverySettlement{TenantId: locked.TenantId, SettlementNo: fmt.Sprintf("%s-%d", batchNo, locked.Id), BatchId: batchID, BatchNo: batchNo, SymbolId: locked.SymbolId, UserId: locked.UserId, PositionId: locked.Id, PositionSide: locked.PositionSide, SettlementPrice: price, PositionQty: locked.Qty, RealizedPnl: pnl, DeliveryFee: fee, SettleAsset: locked.MarginAsset, DeliveryTime: c.DeliveryTime, Status: int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now}); err != nil {
				return err
			}
			locked.Status = int64(trade.PositionStatus_POSITION_STATUS_DELIVERING)
			locked.Version++
			locked.UpdateTimes = now
			if err := pm.Update(ctx, locked); err != nil {
				return err
			}
		}
		symbol.Status = int64(trade.SymbolStatus_SYMBOL_STATUS_DISABLED)
		symbol.UpdateTimes = now
		return symbolModel.Update(ctx, symbol)
	})
}

func (l *ProcessDeliverySettlementsLogic) settlePending(tenantID int64) error {
	for _, status := range []trade.DeliverySettlementStatus{trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_PENDING, trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_FAILED} {
		cursor := int64(0)
		for {
			rows, _, err := l.svcCtx.ContractDeliverySettleModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: tenantID, Status: int64(status)}, cursor, 100)
			if err != nil {
				return err
			}
			if len(rows) == 0 {
				break
			}
			for _, row := range rows {
				cursor = row.Id
				if row.NextRetryAt > utils.NowMillis() {
					continue
				}
				if err := l.settleOne(row); err != nil {
					row.RetryCount++
					row.LastErrorMsg = err.Error()
					row.NextRetryAt = utils.NowMillis() + tradeEventRetryDelay(row.RetryCount).Milliseconds()
					row.Status = int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_FAILED)
					if row.RetryCount >= 20 {
						row.Status = int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_MANUAL_REVIEW)
					}
					row.UpdateTimes = utils.NowMillis()
					if uerr := l.svcCtx.ContractDeliverySettleModel.Update(l.ctx, row); uerr != nil {
						return uerr
					}
				}
			}
			if len(rows) < 100 {
				break
			}
		}
	}
	return l.finishBatches(tenantID)
}

func (l *ProcessDeliverySettlementsLogic) settleOne(row *models.TContractDeliverySettlement) error {
	position, err := l.svcCtx.ContractPositionModel.FindOne(l.ctx, row.PositionId)
	if err != nil {
		return err
	}
	if !position.Qty.IsPositive() || position.Status == int64(trade.PositionStatus_POSITION_STATUS_CLOSED) {
		row.Status = int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_MANUAL_REVIEW)
		row.NextRetryAt = 0
		row.LastErrorMsg = "delivery position was already closed before asset settlement"
		row.UpdateTimes = utils.NowMillis()
		return l.svcCtx.ContractDeliverySettleModel.Update(l.ctx, row)
	}
	calls := []struct {
		suffix string
		credit bool
		amount decimal.Decimal
		remark string
	}{{"MARGIN", true, position.PositionMargin.Add(position.IsolatedMargin), "delivery margin release"}, {"PROFIT", true, decimalMaxZero(row.RealizedPnl), "delivery profit"}, {"LOSS", false, decimalMaxZero(row.RealizedPnl.Neg()), "delivery loss"}, {"FEE", false, row.DeliveryFee, "delivery fee"}}
	for _, call := range calls {
		if !call.amount.IsPositive() {
			continue
		}
		bizNo := row.SettlementNo + "-" + call.suffix
		var resp *asset.ChangeAssetResp
		if call.credit {
			resp, err = l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{TenantId: row.TenantId, UserId: row.UserId, WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: row.SettleAsset, Amount: call.amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: row.Id, BizNo: bizNo, Remark: call.remark})
		} else {
			resp, err = l.svcCtx.AssetClient.SubAvailable(l.ctx, &asset.SubAvailableReq{TenantId: row.TenantId, UserId: row.UserId, WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: row.SettleAsset, Amount: call.amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: row.Id, BizNo: bizNo, Remark: call.remark})
		}
		if err != nil {
			return err
		}
		if resp.GetBase().GetCode() != 200 {
			return fmt.Errorf("delivery asset rejected: %s", resp.GetBase().GetMsg())
		}
	}
	now := utils.NowMillis()
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		pm := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)
		sm := models.NewTContractDeliverySettlementModel(conn, l.svcCtx.Config.CacheRedis)
		hm := models.NewTContractPositionHistoryModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, position.TenantId, position.UserId, position.SymbolId, position.PositionSide, position.MarginMode)
		if err != nil {
			return err
		}
		if current.Qty.IsZero() || current.Status == int64(trade.PositionStatus_POSITION_STATUS_CLOSED) {
			row.Status = int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_MANUAL_REVIEW)
			row.NextRetryAt = 0
			row.LastErrorMsg = "delivery position changed after asset settlement"
			row.UpdateTimes = now
			return sm.Update(ctx, row)
		}
		before := cloneContractPosition(current)
		current.Qty, current.AvailQty, current.FrozenQty = decimal.Zero, decimal.Zero, decimal.Zero
		current.PositionMargin, current.IsolatedMargin, current.MaintenanceMargin = decimal.Zero, decimal.Zero, decimal.Zero
		current.UnrealizedPnl = decimal.Zero
		current.RealizedPnl = current.RealizedPnl.Add(row.RealizedPnl)
		current.Status = int64(trade.PositionStatus_POSITION_STATUS_CLOSED)
		current.ClosedAt = now
		current.Version++
		current.UpdateTimes = now
		if err := pm.Update(ctx, current); err != nil {
			return err
		}
		if err := writeSystemPositionHistory(ctx, hm, before, current, row.SettlementNo, trade.PositionActionType_POSITION_ACTION_TYPE_SETTLEMENT, row.RealizedPnl, row.DeliveryFee, row.SettlementPrice, "delivery settlement"); err != nil {
			return err
		}
		row.Status = int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_SETTLED)
		row.SettledAt = now
		row.NextRetryAt = 0
		row.LastErrorMsg = ""
		row.UpdateTimes = now
		return sm.Update(ctx, row)
	})
}
func (l *ProcessDeliverySettlementsLogic) finishBatches(tenantID int64) error {
	batches, _, err := l.svcCtx.ContractDeliveryBatchModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: tenantID, Status: int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING)}, 0, 100)
	if err != nil {
		return err
	}
	for _, b := range batches {
		count, err := l.svcCtx.ContractDeliverySettleModel.CountByBatchStatus(l.ctx, b.TenantId, b.Id, int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_SETTLED))
		if err != nil {
			return err
		}
		b.SettledPositions = count
		if count == b.TotalPositions {
			b.Status = int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_COMPLETED)
		}
		b.Version++
		b.UpdateTimes = utils.NowMillis()
		if err := l.svcCtx.ContractDeliveryBatchModel.Update(l.ctx, b); err != nil {
			return err
		}
	}
	return nil
}
func normalizeJSON(v any) (string, error) { raw, err := json.Marshal(v); return string(raw), err }
