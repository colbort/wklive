package tradelogic

import (
	"context"
	"errors"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/domain/contractmath"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

// ProcessContractPositionFillsLogic projects derivative Fills into positions.
// Position history action_key is the idempotency boundary for every projection.
type ProcessContractPositionFillsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessContractPositionFillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessContractPositionFillsLogic {
	return &ProcessContractPositionFillsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ProcessContractPositionFillsLogic) ProcessFill(fillID int64) error {
	fill, err := l.svcCtx.TradeFillModel.FindOne(l.ctx, fillID)
	if err != nil {
		return err
	}
	if fill.ProductType != int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		return nil
	}
	order, err := l.svcCtx.TradeOrderModel.FindOne(l.ctx, fill.OrderId)
	if err != nil {
		return err
	}
	contractOrder, err := l.svcCtx.TradeOrderContractModel.FindOneByTenantIdOrderId(l.ctx, fill.TenantId, fill.OrderId)
	if err != nil {
		return err
	}
	contract, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, fill.TenantId, fill.SymbolId)
	if err != nil {
		return err
	}

	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		positionModel := tx.ContractPosition
		historyModel := tx.ContractPositionHistory
		fillModel := tx.TradeFill
		contractOrderModel := tx.TradeOrderContract
		instructionModel := tx.TradeSettlementInstruction
		eventModel := tx.BizTradeEvent
		projected, err := historyModel.CountByRefFillId(ctx, fill.TenantId, fill.Id)
		if err != nil {
			return err
		}
		if projected > 0 {
			return syncFillRealizedPnl(ctx, fillModel, historyModel, fill)
		}

		remaining := fill.Qty
		if fill.PositionSide == int64(trade.PositionSide_POSITION_SIDE_NET) {
			closeSide, openSide := netPositionSides(fill.Side)
			closed, err := l.applyClose(ctx, positionModel, historyModel, eventModel, instructionModel, fill, contractOrder, contract, closeSide, remaining)
			if err != nil {
				return err
			}
			remaining = remaining.Sub(closed)
			if order.IsReduceOnly == int64(common.YesNo_YES_NO_YES) && remaining.IsPositive() {
				return fmt.Errorf("reduce-only fill exceeds position: fill=%s remaining=%s", fill.FillNo, remaining)
			}
			if remaining.IsPositive() {
				if err := l.applyOpen(ctx, positionModel, historyModel, eventModel, instructionModel, fill, contractOrder, contract, openSide, remaining); err != nil {
					return err
				}
			}
		} else if isClosingFill(fill.PositionSide, fill.Side) {
			closed, err := l.applyClose(ctx, positionModel, historyModel, eventModel, instructionModel, fill, contractOrder, contract, fill.PositionSide, remaining)
			if err != nil {
				return err
			}
			if !closed.Equal(remaining) {
				return fmt.Errorf("closing fill exceeds position: fill=%s qty=%s closed=%s", fill.FillNo, remaining, closed)
			}
		} else {
			if err := l.applyOpen(ctx, positionModel, historyModel, eventModel, instructionModel, fill, contractOrder, contract, fill.PositionSide, remaining); err != nil {
				return err
			}
		}

		if contractOrder.ReservedCloseQty.IsPositive() {
			consume := decimalMin(contractOrder.ReservedCloseQty, fill.Qty)
			contractOrder.ReservedCloseQty = contractOrder.ReservedCloseQty.Sub(consume)
			contractOrder.UpdateTimes = utils.NowMillis()
			if err := contractOrderModel.Update(ctx, contractOrder); err != nil {
				return err
			}
		}
		return syncFillRealizedPnl(ctx, fillModel, historyModel, fill)
	})
}

func syncFillRealizedPnl(ctx context.Context, fillModel models.TTradeFillModel, historyModel models.TContractPositionHistoryModel, fill *models.TTradeFill) error {
	histories, err := historyModel.FindByRefFillId(ctx, fill.TenantId, fill.Id)
	if err != nil {
		return err
	}
	realized := decimal.Zero
	for _, history := range histories {
		realized = realized.Add(history.RealizedPnlDelta)
	}
	if fill.RealizedPnl.Equal(realized) {
		return nil
	}
	return fillModel.UpdateRealizedPnl(ctx, fill.Id, realized)
}

func (l *ProcessContractPositionFillsLogic) applyOpen(ctx context.Context, positionModel models.TContractPositionModel, historyModel models.TContractPositionHistoryModel, eventModel models.TBizTradeEventModel, instructionModel models.TTradeSettlementInstructionModel, fill *models.TTradeFill, contractOrder *models.TTradeOrderContract, contract *models.TTradeSymbolContract, side int64, qty decimal.Decimal) error {
	if !qty.IsPositive() {
		return nil
	}
	actionKey := fmt.Sprintf("%s:%d:OPEN", fill.FillNo, side)
	if _, err := historyModel.FindOneByTenantIdActionKey(ctx, fill.TenantId, actionKey); err == nil {
		return nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return err
	}
	position, err := positionModel.FindOneForUpdateByTenantUserSymbolSideMode(ctx, fill.TenantId, fill.UserId, fill.SymbolId, side, contractOrder.MarginMode)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	now := utils.NowMillis()
	positionMode, err := helpers.ContractPositionMode(trade.PositionSide(fill.PositionSide))
	if err != nil {
		return err
	}
	if position == nil {
		position = &models.TContractPosition{TenantId: fill.TenantId, UserId: fill.UserId, SymbolId: fill.SymbolId, ContractType: fill.ContractType, ContractValueType: fill.ContractValueType, PositionSide: side, PositionMode: int64(positionMode), MarginMode: contractOrder.MarginMode, Status: int64(trade.PositionStatus_POSITION_STATUS_NORMAL), Leverage: contractOrder.Leverage, MarginAsset: contractOrder.MarginAsset, CreateTimes: now}
	} else if position.Qty.IsPositive() && position.PositionMode != int64(positionMode) {
		return fmt.Errorf("position mode mismatch: position=%d stored=%d fill=%d", position.Id, position.PositionMode, positionMode)
	} else if position.Qty.IsZero() {
		position.PositionMode = int64(positionMode)
	}
	before := cloneContractPosition(position)
	values, err := contractmath.CalculateTradeValues(fill.ContractValueType, qty, contract.ContractSize, fill.Price)
	if err != nil {
		return err
	}
	margin, err := contractmath.CalculateMargin(values, contractOrder.Leverage)
	if err != nil {
		return err
	}
	position.OpenAvgPrice = contractAveragePrice(position.OpenAvgPrice, position.Qty, fill.Price, qty, fill.ContractValueType)
	position.Qty = position.Qty.Add(qty)
	position.AvailQty = position.AvailQty.Add(qty)
	position.PositionMargin = position.PositionMargin.Add(margin)
	position.MarkPrice = fill.Price
	position.Status = int64(trade.PositionStatus_POSITION_STATUS_NORMAL)
	position.ClosedAt = 0
	riskTier, err := l.riskTierForPosition(ctx, position, contract)
	if err != nil {
		return err
	}
	recalculatePositionRisk(position, contract, riskTier)
	position.Version++
	position.UpdateTimes = now
	if position.Id == 0 {
		result, err := positionModel.Insert(ctx, position)
		if err != nil {
			return err
		}
		position.Id, _ = result.LastInsertId()
	} else if err := positionModel.Update(ctx, position); err != nil {
		return err
	}
	if err := insertSettlementInstructionIdempotent(ctx, instructionModel, &models.TTradeSettlementInstruction{
		TenantId: fill.TenantId, InstructionNo: derivedTradeBizNo(fill.FillNo, "MARGIN"),
		BizType: "fill", BizId: fill.FillNo, BatchNo: fill.MatchNo, FillId: fill.Id,
		OrderId: fill.OrderId, PositionId: position.Id, ReservationNo: fill.OrderNo,
		UserId: fill.UserId, Action: int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_ADJUST_MARGIN),
		Asset: contractOrder.MarginAsset, Amount: contractmath.RoundDebit(margin), StepNo: 1,
		Status:      int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING),
		NextRetryAt: now, CreateTimes: now, UpdateTimes: now,
	}); err != nil {
		return err
	}
	action := trade.PositionActionType_POSITION_ACTION_TYPE_INCREASE
	if before.Qty.IsZero() {
		action = trade.PositionActionType_POSITION_ACTION_TYPE_OPEN
	}
	return writePositionProjection(ctx, historyModel, eventModel, before, position, fill, actionKey, action, decimal.Zero, qty)
}

func (l *ProcessContractPositionFillsLogic) applyClose(ctx context.Context, positionModel models.TContractPositionModel, historyModel models.TContractPositionHistoryModel, eventModel models.TBizTradeEventModel, instructionModel models.TTradeSettlementInstructionModel, fill *models.TTradeFill, contractOrder *models.TTradeOrderContract, contract *models.TTradeSymbolContract, side int64, requested decimal.Decimal) (decimal.Decimal, error) {
	if !requested.IsPositive() {
		return decimal.Zero, nil
	}
	actionKey := fmt.Sprintf("%s:%d:CLOSE", fill.FillNo, side)
	if history, err := historyModel.FindOneByTenantIdActionKey(ctx, fill.TenantId, actionKey); err == nil {
		return history.BeforeQty.Sub(history.AfterQty), nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return decimal.Zero, err
	}
	position, err := positionModel.FindOneForUpdateByTenantUserSymbolSideMode(ctx, fill.TenantId, fill.UserId, fill.SymbolId, side, contractOrder.MarginMode)
	if errors.Is(err, models.ErrNotFound) {
		return decimal.Zero, nil
	}
	if err != nil {
		return decimal.Zero, err
	}
	positionMode, err := helpers.ContractPositionMode(trade.PositionSide(fill.PositionSide))
	if err != nil {
		return decimal.Zero, err
	}
	if position.PositionMode != int64(positionMode) {
		return decimal.Zero, fmt.Errorf("position mode mismatch: position=%d stored=%d fill=%d", position.Id, position.PositionMode, positionMode)
	}
	qty := decimalMin(position.Qty, requested)
	if !qty.IsPositive() {
		return decimal.Zero, nil
	}
	before := cloneContractPosition(position)
	realized := contractRealizedPnl(side, position.OpenAvgPrice, fill.Price, qty, contract.ContractSize, fill.ContractValueType)
	marginReleased := proportionalAmount(position.PositionMargin, qty, position.Qty)
	position.Qty = position.Qty.Sub(qty)
	position.PositionMargin = decimalMaxZero(position.PositionMargin.Sub(marginReleased))
	if contractOrder.ReservedCloseQty.IsPositive() {
		position.FrozenQty = decimalMaxZero(position.FrozenQty.Sub(qty))
	} else {
		position.AvailQty = decimalMaxZero(position.AvailQty.Sub(qty))
	}
	position.RealizedPnl = position.RealizedPnl.Add(realized)
	position.MarkPrice = fill.Price
	if position.Qty.IsZero() {
		position.AvailQty, position.FrozenQty, position.OpenAvgPrice = decimal.Zero, decimal.Zero, decimal.Zero
		position.PositionMargin, position.MaintenanceMargin, position.UnrealizedPnl = decimal.Zero, decimal.Zero, decimal.Zero
		position.LiquidationPrice, position.BankruptcyPrice, position.RiskRate = decimal.Zero, decimal.Zero, decimal.Zero
		position.Status = int64(trade.PositionStatus_POSITION_STATUS_CLOSED)
		position.ClosedAt = utils.NowMillis()
	} else {
		riskTier, err := l.riskTierForPosition(ctx, position, contract)
		if err != nil {
			return decimal.Zero, err
		}
		recalculatePositionRisk(position, contract, riskTier)
	}
	position.Version++
	position.UpdateTimes = utils.NowMillis()
	if err := positionModel.Update(ctx, position); err != nil {
		return decimal.Zero, err
	}
	action := trade.PositionActionType_POSITION_ACTION_TYPE_DECREASE
	if position.Qty.IsZero() {
		action = trade.PositionActionType_POSITION_ACTION_TYPE_CLOSE
	}
	if err := writePositionProjection(ctx, historyModel, eventModel, before, position, fill, actionKey, action, realized, qty); err != nil {
		return decimal.Zero, err
	}
	if err := writeContractCloseSettlements(ctx, instructionModel, fill, contractOrder, position.Id, marginReleased, realized); err != nil {
		return decimal.Zero, err
	}
	return qty, nil
}

func writeContractCloseSettlements(ctx context.Context, instructionModel models.TTradeSettlementInstructionModel, fill *models.TTradeFill, contractOrder *models.TTradeOrderContract, positionID int64, marginReleased, realized decimal.Decimal) error {
	now := utils.NowMillis()
	stepNo := int64(10)
	insert := func(suffix string, action trade.SettlementInstructionAction, amount decimal.Decimal) error {
		if !amount.IsPositive() {
			return nil
		}
		instruction := &models.TTradeSettlementInstruction{TenantId: fill.TenantId, InstructionNo: derivedTradeBizNo(fill.FillNo, suffix), BizType: "fill", BizId: fill.FillNo, BatchNo: fill.MatchNo, FillId: fill.Id, OrderId: fill.OrderId, PositionId: positionID, ReservationNo: fill.OrderNo, UserId: fill.UserId, Action: int64(action), Asset: contractOrder.MarginAsset, Amount: amount, StepNo: stepNo, Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now}
		stepNo++
		return insertSettlementInstructionIdempotent(ctx, instructionModel, instruction)
	}
	if err := insert("MARGIN_RELEASE", trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_MARGIN, contractmath.RoundCredit(marginReleased)); err != nil {
		return err
	}
	if realized.IsPositive() {
		return insert("PNL_PROFIT", trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_POST_PNL, contractmath.RoundCredit(realized))
	}
	if realized.IsNegative() {
		return insert("PNL_LOSS", trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS, contractmath.RoundDebit(realized.Abs()))
	}
	return nil
}

func (l *ProcessContractPositionFillsLogic) riskTierForPosition(ctx context.Context, position *models.TContractPosition, contract *models.TTradeSymbolContract) (*models.TContractRiskLimitTier, error) {
	values, err := contractmath.CalculateTradeValues(position.ContractValueType, position.Qty, contract.ContractSize, position.MarkPrice)
	if err != nil {
		return nil, err
	}
	tier, err := l.svcCtx.ContractRiskLimitTierModel.FindByNotional(ctx, position.TenantId, position.SymbolId, values.QuoteNotional)
	if errors.Is(err, models.ErrNotFound) {
		return nil, nil
	}
	return tier, err
}

func writePositionProjection(ctx context.Context, historyModel models.TContractPositionHistoryModel, eventModel models.TBizTradeEventModel, before, after *models.TContractPosition, fill *models.TTradeFill, actionKey string, action trade.PositionActionType, realized, appliedQty decimal.Decimal) error {
	now := utils.NowMillis()
	feeDelta := proportionalAmount(fill.Fee, appliedQty, fill.Qty)
	if _, err := historyModel.Insert(ctx, &models.TContractPositionHistory{TenantId: after.TenantId, PositionId: after.Id, UserId: after.UserId, SymbolId: after.SymbolId, ContractType: after.ContractType, ContractValueType: after.ContractValueType, PositionSide: after.PositionSide, MarginAsset: after.MarginAsset, ActionType: int64(action), ActionKey: actionKey, BusinessTime: fill.CreateTimes, BeforeVersion: before.Version, AfterVersion: after.Version, BeforeQty: before.Qty, AfterQty: after.Qty, BeforeAvailQty: before.AvailQty, AfterAvailQty: after.AvailQty, BeforeFrozenQty: before.FrozenQty, AfterFrozenQty: after.FrozenQty, BeforeOpenAvgPrice: before.OpenAvgPrice, AfterOpenAvgPrice: after.OpenAvgPrice, BeforePositionMargin: before.PositionMargin, AfterPositionMargin: after.PositionMargin, BeforeIsolatedMargin: before.IsolatedMargin, AfterIsolatedMargin: after.IsolatedMargin, BeforeUnrealizedPnl: before.UnrealizedPnl, AfterUnrealizedPnl: after.UnrealizedPnl, RealizedPnlDelta: realized, FeeDelta: feeDelta, FeeAsset: fill.FeeAsset, MarkPrice: fill.Price, RefOrderId: fill.OrderId, RefFillId: fill.Id, Source: int64(trade.SourceType_SOURCE_TYPE_SYSTEM), Remark: "position projected from fill", CreateTimes: now}); err != nil {
		return err
	}
	eventNo := derivedTradeBizNo(fill.FillNo, fmt.Sprintf("POS%d", after.PositionSide))
	_, err := eventModel.Insert(ctx, &models.TBizTradeEvent{TenantId: after.TenantId, EventNo: eventNo, EventType: "POSITION_UPDATED", BizId: fmt.Sprintf("%d", after.Id), BizType: "position", UserId: after.UserId, SymbolId: after.SymbolId, ProductType: int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE), Source: int64(trade.SourceType_SOURCE_TYPE_SYSTEM), EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20, NextRetryAt: now, Payload: "{}", CreateTimes: now, UpdateTimes: now})
	return err
}

func netPositionSides(side int64) (closeSide, openSide int64) {
	if side == int64(common.Side_SIDE_BUY) {
		return int64(trade.PositionSide_POSITION_SIDE_SHORT), int64(trade.PositionSide_POSITION_SIDE_LONG)
	}
	return int64(trade.PositionSide_POSITION_SIDE_LONG), int64(trade.PositionSide_POSITION_SIDE_SHORT)
}

func isClosingFill(positionSide, side int64) bool {
	return positionSide == int64(trade.PositionSide_POSITION_SIDE_LONG) && side == int64(common.Side_SIDE_SELL) || positionSide == int64(trade.PositionSide_POSITION_SIDE_SHORT) && side == int64(common.Side_SIDE_BUY)
}

func contractAveragePrice(oldPrice, oldQty, fillPrice, fillQty decimal.Decimal, valueType int64) decimal.Decimal {
	if oldQty.IsZero() {
		return fillPrice
	}
	total := oldQty.Add(fillQty)
	if valueType == int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE) {
		return total.Div(oldQty.Div(oldPrice).Add(fillQty.Div(fillPrice)))
	}
	return oldPrice.Mul(oldQty).Add(fillPrice.Mul(fillQty)).Div(total)
}

func contractRealizedPnl(side int64, openPrice, closePrice, qty, contractSize decimal.Decimal, valueType int64) decimal.Decimal {
	contracts := qty.Mul(contractSize)
	if valueType == int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE) {
		pnl := contracts.Mul(decimal.NewFromInt(1).Div(openPrice).Sub(decimal.NewFromInt(1).Div(closePrice)))
		if side == int64(trade.PositionSide_POSITION_SIDE_SHORT) {
			return contractmath.RoundCredit(pnl.Neg())
		}
		return contractmath.RoundCredit(pnl)
	}
	pnl := closePrice.Sub(openPrice).Mul(contracts)
	if side == int64(trade.PositionSide_POSITION_SIDE_SHORT) {
		return contractmath.RoundCredit(pnl.Neg())
	}
	return contractmath.RoundCredit(pnl)
}

func recalculatePositionRisk(position *models.TContractPosition, contract *models.TTradeSymbolContract, tiers ...*models.TContractRiskLimitTier) {
	position.UnrealizedPnl = contractRealizedPnl(position.PositionSide, position.OpenAvgPrice, position.MarkPrice, position.Qty, contract.ContractSize, position.ContractValueType)
	values, err := contractmath.CalculateTradeValues(position.ContractValueType, position.Qty, contract.ContractSize, position.MarkPrice)
	if err != nil {
		position.Status = int64(trade.PositionStatus_POSITION_STATUS_MANUAL_REVIEW)
		return
	}
	maintenanceRate := contract.MaintenanceMarginRate
	maintenanceAmount := decimal.Zero
	if len(tiers) > 0 && tiers[0] != nil {
		maintenanceRate = tiers[0].MaintenanceMarginRate
		maintenanceAmount = tiers[0].MaintenanceAmount
	}
	position.MaintenanceMargin = decimalMaxZero(contractmath.RoundDebit(values.SettlementNotional.Mul(maintenanceRate)).Sub(maintenanceAmount))
	if position.MarginMode == int64(trade.MarginMode_MARGIN_MODE_CROSS) {
		// Cross liquidation is account-level and requires wallet equity plus all
		// positions sharing the margin asset. Never persist an isolated-position
		// approximation that could trigger an incorrect liquidation.
		position.RiskRate = decimal.Zero
		position.BankruptcyPrice = decimal.Zero
		position.LiquidationPrice = decimal.Zero
		return
	}
	equity := position.PositionMargin.Add(position.IsolatedMargin).Add(position.UnrealizedPnl)
	if equity.IsPositive() {
		position.RiskRate = position.MaintenanceMargin.Div(equity)
	} else {
		position.RiskRate = decimal.Zero
	}
	position.BankruptcyPrice, position.LiquidationPrice = contractRiskPricesWithMaintenance(position, contract, maintenanceRate, maintenanceAmount)
}

func contractRiskPrices(position *models.TContractPosition, contract *models.TTradeSymbolContract) (decimal.Decimal, decimal.Decimal) {
	return contractRiskPricesWithMaintenance(position, contract, contract.MaintenanceMarginRate, decimal.Zero)
}

func contractRiskPricesWithMaintenance(position *models.TContractPosition, contract *models.TTradeSymbolContract, maintenanceRate, maintenanceAmount decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	contracts := position.Qty.Mul(contract.ContractSize)
	if !contracts.IsPositive() || !position.OpenAvgPrice.IsPositive() {
		return decimal.Zero, decimal.Zero
	}
	margin := position.PositionMargin.Add(position.IsolatedMargin)
	if position.ContractValueType == int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE) {
		base := decimal.NewFromInt(1).Div(position.OpenAvgPrice)
		bankDenom := base
		if position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_LONG) {
			bankDenom = base.Add(margin.Div(contracts))
			liquidation := contracts.Mul(decimal.NewFromInt(1).Add(maintenanceRate)).Div(margin.Add(contracts.Mul(base)).Add(maintenanceAmount))
			return decimal.NewFromInt(1).Div(bankDenom), liquidation
		} else {
			bankDenom = base.Sub(margin.Div(contracts))
		}
		liqDenom := contracts.Mul(base).Sub(margin).Sub(maintenanceAmount)
		if !bankDenom.IsPositive() || !liqDenom.IsPositive() {
			return decimal.Zero, decimal.Zero
		}
		liquidation := contracts.Mul(decimal.NewFromInt(1).Sub(maintenanceRate)).Div(liqDenom)
		return decimal.NewFromInt(1).Div(bankDenom), liquidation
	}
	bankDelta := margin.Div(contracts)
	if position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_LONG) {
		bankruptcy := decimalMaxZero(position.OpenAvgPrice.Sub(bankDelta))
		denominator := decimal.NewFromInt(1).Sub(maintenanceRate)
		if !denominator.IsPositive() {
			return bankruptcy, decimal.Zero
		}
		liquidation := decimalMaxZero(position.OpenAvgPrice.Sub(bankDelta).Sub(maintenanceAmount.Div(contracts)).Div(denominator))
		return bankruptcy, liquidation
	}
	bankruptcy := position.OpenAvgPrice.Add(bankDelta)
	liquidation := position.OpenAvgPrice.Add(bankDelta).Add(maintenanceAmount.Div(contracts)).Div(decimal.NewFromInt(1).Add(maintenanceRate))
	return bankruptcy, liquidation
}

func proportionalAmount(amount, part, total decimal.Decimal) decimal.Decimal {
	if !total.IsPositive() {
		return decimal.Zero
	}
	return amount.Mul(part).Div(total)
}

func cloneContractPosition(position *models.TContractPosition) *models.TContractPosition {
	copy := *position
	return &copy
}

func decimalMin(a, b decimal.Decimal) decimal.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}

func decimalMaxZero(value decimal.Decimal) decimal.Decimal {
	if value.IsNegative() {
		return decimal.Zero
	}
	return value
}
