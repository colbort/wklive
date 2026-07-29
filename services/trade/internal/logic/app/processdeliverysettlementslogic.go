package applogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"wklive/proto/common"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/domain/contractmath"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessDeliverySettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProcessDeliverySettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessDeliverySettlementsLogic {
	return &ProcessDeliverySettlementsLogic{ctx: ctx, svcCtx: svcCtx}
}

func validateFinalDeliveryPriceFact(configuredAlgorithm string, quote *marketQuoteSnapshot, candidates []*marketQuoteSnapshot) (string, string, error) {
	if quote == nil || quote.SnapshotID == "" || !quote.Confirmed {
		return "", "", errors.New("final DELIVERY snapshot is missing or unconfirmed")
	}
	if len(candidates) != 1 || candidates[0] == nil || candidates[0].SnapshotID != quote.SnapshotID {
		return "", "", fmt.Errorf("Trade requires exactly one final Price Engine DELIVERY snapshot, got %d candidates", len(candidates))
	}
	formulaVersion := strings.TrimSpace(quote.FormulaVersion)
	if formulaVersion == "" {
		return "", "", errors.New("final DELIVERY snapshot has no formula version")
	}
	if strings.TrimSpace(configuredAlgorithm) == "" {
		return "", "", errors.New("delivery settlement price algorithm is not configured")
	}
	return strings.TrimSpace(configuredAlgorithm), formulaVersion, nil
}

func (l *ProcessDeliverySettlementsLogic) Process(tenantID int64) error {
	if err := l.advanceSymbols(tenantID); err != nil {
		return err
	}
	if err := l.settlePending(tenantID); err != nil {
		return err
	}
	return l.archiveCompletedBatches(tenantID)
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
			if symbol.ContractType != int64(common.ContractType_CONTRACT_TYPE_DELIVERY) {
				continue
			}
			if c.OpenCutoffTime > 0 && now >= c.OpenCutoffTime && symbol.Status == int64(trade.SymbolStatus_SYMBOL_STATUS_ENABLED) {
				symbol.Status = int64(trade.SymbolStatus_SYMBOL_STATUS_CLOSE_ONLY)
				symbol.UpdateTimes = now
				if err := l.svcCtx.TradeSymbolModel.Update(l.ctx, symbol); err != nil {
					return err
				}
			}
			if c.OpenCutoffTime > 0 && now >= c.OpenCutoffTime {
				if err := l.ensureLifecycleBatch(c, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_CLOSE_ONLY); err != nil {
					return err
				}
			}
			if c.MatchingStopTime > 0 && now >= c.MatchingStopTime {
				if err := l.cancelSymbolOrders(symbol); err != nil {
					return err
				}
				if err := l.ensureLifecycleBatch(c, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_MATCHING_STOPPED); err != nil {
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
			if err := l.ensureLifecycleBatch(c, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_PRICE_LOCKING); err != nil {
				return err
			}
			if err := l.ensureBatch(symbol, c); err != nil {
				_ = l.recordDeliveryBatchError(c, err)
				return err
			}
		}
		if len(contracts) < 100 {
			return nil
		}
	}
}

func (l *ProcessDeliverySettlementsLogic) ensureLifecycleBatch(c *models.TTradeSymbolContract, target trade.DeliveryBatchStatus) error {
	if c == nil || c.DeliveryTime <= 0 {
		return errors.New("delivery lifecycle requires a delivery contract")
	}
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractDeliveryBatch
		current, err := bm.FindOneForUpdateByTenantSymbolDelivery(ctx, c.TenantId, c.SymbolId, c.DeliveryTime)
		if errors.Is(err, models.ErrNotFound) {
			_, err = bm.Insert(ctx, &models.TContractDeliveryBatch{
				TenantId: c.TenantId, BatchNo: fmt.Sprintf("DLV-%d-%d", c.SymbolId, c.DeliveryTime),
				SymbolId: c.SymbolId, OpenCutoffTime: c.OpenCutoffTime, MatchingStopTime: c.MatchingStopTime,
				DeliveryTime: c.DeliveryTime, Status: int64(target), CreateTimes: now, UpdateTimes: now,
			})
			return err
		}
		if err != nil {
			return err
		}
		// Lifecycle states are monotonic. Once settlement has started, an
		// earlier scheduler stage must never move the batch backwards.
		if !shouldAdvanceDeliveryBatchStatus(current.Status, target) {
			return nil
		}
		current.Status = int64(target)
		current.Version++
		current.UpdateTimes = now
		return bm.Update(ctx, current)
	})
}

func shouldAdvanceDeliveryBatchStatus(current int64, target trade.DeliveryBatchStatus) bool {
	targetStatus := int64(target)
	if targetStatus < int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_CLOSE_ONLY) ||
		targetStatus > int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_PRICE_LOCKING) {
		return false
	}
	return current < targetStatus &&
		current < int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING)
}

func (l *ProcessDeliverySettlementsLogic) recordDeliveryBatchError(c *models.TTradeSymbolContract, cause error) error {
	if c == nil || cause == nil {
		return nil
	}
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractDeliveryBatch
		current, err := bm.FindOneForUpdateByTenantSymbolDelivery(ctx, c.TenantId, c.SymbolId, c.DeliveryTime)
		if err != nil {
			return err
		}
		current.LastErrorMsg = cause.Error()
		current.Version++
		current.UpdateTimes = now
		return bm.Update(ctx, current)
	})
}

func (l *ProcessDeliverySettlementsLogic) cancelSymbolOrders(symbol *models.TTradeSymbol) error {
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{TenantId: symbol.TenantId, SymbolId: symbol.Id, ProductType: int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE), Statuses: []int64{int64(trade.OrderStatus_ORDER_STATUS_PENDING), int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED), int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING)}}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, o := range orders {
			cursor = o.Id
			terminating, err := beginSystemOrderTermination(l.ctx, l.svcCtx, o.Id, "delivery matching stopped", false)
			if err != nil {
				return err
			}
			if terminating == nil {
				continue
			}
			if err := unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, terminating, "delivery order release"); err != nil {
				return err
			}
			if err := removeOrderBookOrder(l.svcCtx, l.ctx, terminating); err != nil {
				logx.WithContext(l.ctx).Errorf("remove delivery order from cache failed, orderId=%d err=%v", terminating.Id, err)
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessDeliverySettlementsLogic) ensureBatch(symbol *models.TTradeSymbol, c *models.TTradeSymbolContract) error {
	existing, err := l.svcCtx.ContractDeliveryBatchModel.FindOneByTenantIdSymbolIdDeliveryTime(l.ctx, c.TenantId, c.SymbolId, c.DeliveryTime)
	if err == nil && existing.Status >= int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING) {
		return nil
	}
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	quote, candidates, err := NewProcessSecondsSettlementsLogic(l.ctx, l.svcCtx).getValidQuotesAtKind("DELIVERY_PRICE", c.SettlementPriceSource, c.SymbolId, c.DeliveryTime, c.SettlementWindowSeconds*1000)
	if err != nil {
		return err
	}
	priceAlgorithm, formulaVersion, err := validateFinalDeliveryPriceFact(c.SettlementPriceAlgorithm, quote, candidates)
	if err != nil {
		return err
	}
	window := c.SettlementWindowSeconds * 1000
	if window <= 0 || quote.QuoteTs < c.DeliveryTime-window || quote.QuoteTs > c.DeliveryTime+window {
		return fmt.Errorf("delivery quote outside configured settlement window")
	}
	price := helpers.MustParseFloat(quote.LastPrice)
	positions, err := l.svcCtx.ContractPositionModel.FindList(l.ctx, models.ContractPositionPageFilter{TenantId: c.TenantId, SymbolId: c.SymbolId, ContractType: int64(common.ContractType_CONTRACT_TYPE_DELIVERY)})
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
	raw, _ := normalizeJSON(map[string]any{"quote": quote, "configured_algorithm": c.SettlementPriceAlgorithm, "trade_policy": priceAlgorithm})
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractDeliveryBatch
		sm := tx.ContractDeliverySettlement
		im := tx.TradeSettlementInstruction
		pm := tx.ContractPosition
		symbolModel := tx.TradeSymbol
		currentBatch, err := bm.FindOneForUpdateByTenantSymbolDelivery(ctx, c.TenantId, c.SymbolId, c.DeliveryTime)
		var batchID int64
		if errors.Is(err, models.ErrNotFound) {
			res, insertErr := bm.Insert(ctx, &models.TContractDeliveryBatch{TenantId: c.TenantId, BatchNo: batchNo, SymbolId: c.SymbolId, SettlementPrice: price, PriceSource: quote.SnapshotID, PriceAlgorithm: priceAlgorithm, FormulaVersion: formulaVersion, SampleSnapshot: sql.NullString{String: raw, Valid: true}, OpenCutoffTime: c.OpenCutoffTime, MatchingStopTime: c.MatchingStopTime, DeliveryTime: c.DeliveryTime, Status: int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING), TotalPositions: int64(len(active)), CreateTimes: now, UpdateTimes: now})
			if insertErr != nil {
				return insertErr
			}
			batchID, _ = res.LastInsertId()
		} else {
			if err != nil {
				return err
			}
			if currentBatch.Status >= int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING) {
				return nil
			}
			currentBatch.SettlementPrice = price
			currentBatch.PriceSource = quote.SnapshotID
			currentBatch.PriceAlgorithm = priceAlgorithm
			currentBatch.FormulaVersion = formulaVersion
			currentBatch.SampleSnapshot = sql.NullString{String: raw, Valid: true}
			currentBatch.Status = int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING)
			currentBatch.TotalPositions = int64(len(active))
			currentBatch.LastErrorMsg = ""
			currentBatch.Version++
			currentBatch.UpdateTimes = now
			if err = bm.Update(ctx, currentBatch); err != nil {
				return err
			}
			batchID = currentBatch.Id
		}
		for _, p := range active {
			locked, err := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, p.TenantId, p.UserId, p.SymbolId, p.PositionSide, p.MarginMode)
			if err != nil {
				return err
			}
			if !locked.Qty.Equal(p.Qty) {
				return fmt.Errorf("delivery position changed during price lock: %d", p.Id)
			}
			pnl := contractRealizedPnl(locked.PositionSide, locked.OpenAvgPrice, price, locked.Qty, c.ContractSize, locked.ContractValueType)
			values, err := contractmath.CalculateTradeValues(locked.ContractValueType, locked.Qty, c.ContractSize, price)
			if err != nil {
				return err
			}
			fee := contractmath.RoundDebit(values.SettlementNotional.Mul(c.DeliveryFeeRate))
			settlementNo := fmt.Sprintf("%s-%d", batchNo, locked.Id)
			steps := deliveryAssetSteps(locked.PositionMargin.Add(locked.IsolatedMargin), pnl, fee)
			settlement := &models.TContractDeliverySettlement{TenantId: locked.TenantId, SettlementNo: settlementNo, BatchId: batchID, BatchNo: batchNo, SymbolId: locked.SymbolId, UserId: locked.UserId, PositionId: locked.Id, PositionSide: locked.PositionSide, SettlementPrice: price, PositionQty: locked.Qty, RealizedPnl: pnl, DeliveryFee: fee, SettleAsset: locked.MarginAsset, DeliveryTime: c.DeliveryTime, Status: int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now}
			if len(steps) == 0 {
				settlement.Status = int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_SETTLED)
				settlement.SettledAt = now
				settlement.NextRetryAt = 0
			}
			if _, err = sm.Insert(ctx, settlement); err != nil {
				return err
			}
			for _, step := range steps {
				if err = insertSettlementInstructionIdempotent(ctx, im, &models.TTradeSettlementInstruction{TenantId: locked.TenantId, InstructionNo: settlementNo + "-" + step.suffix, BizType: "delivery", BizId: settlementNo, BatchNo: batchNo, PositionId: locked.Id, UserId: locked.UserId, Action: int64(step.action), Asset: locked.MarginAsset, Amount: step.amount, StepNo: step.stepNo, Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now}); err != nil {
					return err
				}
			}
			before := cloneContractPosition(locked)
			locked.Status = int64(trade.PositionStatus_POSITION_STATUS_DELIVERING)
			if len(steps) == 0 {
				locked.Qty, locked.AvailQty, locked.FrozenQty = decimal.Zero, decimal.Zero, decimal.Zero
				locked.PositionMargin, locked.IsolatedMargin, locked.MaintenanceMargin = decimal.Zero, decimal.Zero, decimal.Zero
				locked.UnrealizedPnl = decimal.Zero
				locked.RealizedPnl = locked.RealizedPnl.Add(pnl)
				locked.Status = int64(trade.PositionStatus_POSITION_STATUS_CLOSED)
				locked.ClosedAt = now
				locked.Version++
				locked.UpdateTimes = now
				if err = writeSystemPositionHistory(ctx, tx.ContractPositionHistory, before, locked, c.DeliveryTime, settlementNo, trade.PositionActionType_POSITION_ACTION_TYPE_SETTLEMENT, pnl, fee, price, "delivery settlement without asset step"); err != nil {
					return err
				}
			} else {
				locked.Version++
				locked.UpdateTimes = now
			}
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
	for processed := 0; processed < 1000; {
		now := utils.NowMillis()
		items, err := l.svcCtx.TradeSettlementInstrModel.FindPendingBiz(l.ctx, tenantID, "delivery", now, 100)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			break
		}
		progressed := false
		for _, item := range items {
			claimed, lease, claimErr := l.svcCtx.TradeSettlementInstrModel.ClaimLease(l.ctx, item.Id, now)
			if claimErr != nil {
				return claimErr
			}
			if !claimed {
				continue
			}
			item.UpdateTimes = lease
			progressed = true
			processed++
			if executeErr := l.executeDeliveryInstruction(item); executeErr != nil {
				if failErr := l.failDeliveryInstruction(item, executeErr); failErr != nil {
					return failErr
				}
			}
		}
		if !progressed {
			break
		}
	}
	return l.finishBatches(tenantID)
}

func (l *ProcessDeliverySettlementsLogic) executeDeliveryInstruction(item *models.TTradeSettlementInstruction) error {
	row, err := l.svcCtx.ContractDeliverySettleModel.FindOneByTenantIdSettlementNo(l.ctx, item.TenantId, item.BizId)
	if err != nil {
		return err
	}
	if row.BatchNo != item.BatchNo || row.PositionId != item.PositionId || row.UserId != item.UserId || row.SettleAsset != item.Asset {
		return errors.New("delivery instruction does not match settlement")
	}
	position, err := l.svcCtx.ContractPositionModel.FindOne(l.ctx, row.PositionId)
	if err != nil {
		return err
	}
	if position.Status != int64(trade.PositionStatus_POSITION_STATUS_DELIVERING) || !position.Qty.Equal(row.PositionQty) {
		return errors.New("delivery reserved position changed before asset step")
	}
	margin := position.PositionMargin.Add(position.IsolatedMargin)
	if !matchesDeliveryAssetStep(item, deliveryAssetSteps(margin, row.RealizedPnl, row.DeliveryFee)) &&
		!matchesDeliveryAssetStep(item, legacyDeliveryAssetSteps(margin, row.RealizedPnl, row.DeliveryFee)) {
		return errors.New("delivery instruction action, amount or step was modified")
	}
	if err = executeSimpleAssetInstruction(l.ctx, l.svcCtx, item, "delivery settlement"); err != nil {
		return err
	}
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		im := tx.TradeSettlementInstruction
		currentInstruction, err := im.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if currentInstruction.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			return nil
		}
		if !settlementInstructionLeaseOwned(currentInstruction, item) {
			return errors.New("delivery instruction lease lost")
		}
		currentInstruction.Status, currentInstruction.NextRetryAt, currentInstruction.LastErrorMsg, currentInstruction.UpdateTimes = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS), 0, "", now
		if err = im.Update(ctx, currentInstruction); err != nil {
			return err
		}
		unfinished, err := im.CountUnfinishedByBiz(ctx, item.TenantId, "delivery", item.BizId)
		if err != nil || unfinished > 0 {
			return err
		}
		pm := tx.ContractPosition
		sm := tx.ContractDeliverySettlement
		hm := tx.ContractPositionHistory
		current, err := pm.FindOneForUpdateByTenantUserSymbolSideMode(ctx, position.TenantId, position.UserId, position.SymbolId, position.PositionSide, position.MarginMode)
		if err != nil {
			return err
		}
		if current.Status != int64(trade.PositionStatus_POSITION_STATUS_DELIVERING) || !current.Qty.Equal(row.PositionQty) {
			return errors.New("delivery reserved position changed after asset step")
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
		if err := writeSystemPositionHistory(ctx, hm, before, current, row.DeliveryTime, row.SettlementNo, trade.PositionActionType_POSITION_ACTION_TYPE_SETTLEMENT, row.RealizedPnl, row.DeliveryFee, row.SettlementPrice, "delivery settlement"); err != nil {
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

func (l *ProcessDeliverySettlementsLogic) failDeliveryInstruction(item *models.TTradeSettlementInstruction, cause error) error {
	return failContractSagaInstruction(l.ctx, l.svcCtx, item, cause, func(ctx context.Context, tx *models.TransactionModels, current *models.TTradeSettlementInstruction, manual bool, now int64) error {
		sm := tx.ContractDeliverySettlement
		row, err := sm.FindOneByTenantIdSettlementNo(ctx, current.TenantId, current.BizId)
		if err != nil {
			return err
		}
		row.Status = int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_FAILED)
		if manual {
			row.Status = int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_MANUAL_REVIEW)
		}
		row.RetryCount, row.NextRetryAt, row.LastErrorMsg, row.UpdateTimes = current.RetryCount, current.NextRetryAt, current.LastErrorMsg, now
		return sm.Update(ctx, row)
	})
}
func (l *ProcessDeliverySettlementsLogic) finishBatches(tenantID int64) error {
	batches, _, err := l.svcCtx.ContractDeliveryBatchModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: tenantID, Status: int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING)}, 0, 100)
	if err != nil {
		return err
	}
	for _, b := range batches {
		unfinished, instructionErr := l.svcCtx.TradeSettlementInstrModel.CountUnfinishedByBatch(l.ctx, b.TenantId, "delivery", b.BatchNo)
		if instructionErr != nil {
			return instructionErr
		}
		unreconciled, reconcileErr := l.svcCtx.TradeSettlementInstrModel.CountUnreconciledByBatch(l.ctx, b.TenantId, "delivery", b.BatchNo)
		if reconcileErr != nil {
			return reconcileErr
		}
		manual, manualErr := l.svcCtx.TradeSettlementInstrModel.CountByBatchStatus(l.ctx, b.TenantId, "delivery", b.BatchNo, int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW))
		if manualErr != nil {
			return manualErr
		}
		count, err := l.svcCtx.ContractDeliverySettleModel.CountByBatchStatus(l.ctx, b.TenantId, b.Id, int64(trade.DeliverySettlementStatus_DELIVERY_SETTLEMENT_STATUS_SETTLED))
		if err != nil {
			return err
		}
		b.SettledPositions = count
		if manual > 0 {
			b.Status = int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_MANUAL_REVIEW)
			b.LastErrorMsg = "one or more delivery asset instructions require manual review"
		} else if count == b.TotalPositions && unfinished == 0 && unreconciled == 0 {
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

func (l *ProcessDeliverySettlementsLogic) archiveCompletedBatches(tenantID int64) error {
	batches, _, err := l.svcCtx.ContractDeliveryBatchModel.FindPage(l.ctx, models.AdminPageFilter{
		TenantId: tenantID,
		Status:   int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_COMPLETED),
	}, 0, 100)
	if err != nil {
		return err
	}
	for _, batch := range batches {
		now := utils.NowMillis()
		if err = l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
			batchModel := tx.ContractDeliveryBatch
			eventModel := tx.BizTradeEvent
			current, findErr := batchModel.FindOne(ctx, batch.Id)
			if findErr != nil {
				return findErr
			}
			if current.Status != int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_COMPLETED) {
				return nil
			}
			current.Status = int64(trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_ARCHIVED)
			current.Version++
			current.UpdateTimes = now
			if updateErr := batchModel.Update(ctx, current); updateErr != nil {
				return updateErr
			}
			_, insertErr := eventModel.Insert(ctx, &models.TBizTradeEvent{
				TenantId:      current.TenantId,
				EventNo:       current.BatchNo + "-ARCHIVED",
				EventType:     "CONTRACT_SETTLED",
				BizId:         current.BatchNo,
				BizType:       "delivery_batch",
				SymbolId:      current.SymbolId,
				ProductType:   int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE),
				Source:        int64(trade.SourceType_SOURCE_TYPE_TASK),
				EventStatus:   int64(trade.EventStatus_EVENT_STATUS_PENDING),
				MaxRetryCount: 20,
				NextRetryAt:   now,
				Payload:       helpers.NormalizeTradeEventJSON(current.BatchNo),
				CreateTimes:   now,
				UpdateTimes:   now,
			})
			return insertErr
		}); err != nil {
			return err
		}
	}
	return nil
}
func normalizeJSON(v any) (string, error) { raw, err := json.Marshal(v); return string(raw), err }
