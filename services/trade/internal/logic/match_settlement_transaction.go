package logic

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

type fillEventPayload struct {
	MatchNo      string `json:"match_no"`
	FillNo       string `json:"fill_no"`
	OrderNo      string `json:"order_no"`
	OrderID      int64  `json:"order_id"`
	FillID       int64  `json:"fill_id"`
	UserID       int64  `json:"user_id"`
	SymbolID     int64  `json:"symbol_id"`
	ProductType  int64  `json:"product_type"`
	Side         int64  `json:"side"`
	PositionSide int64  `json:"position_side"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	Amount       string `json:"amount"`
	Fee          string `json:"fee"`
	FeeAsset     string `json:"fee_asset"`
	OrderStatus  int64  `json:"order_status"`
}

type settlementInstructionSpec struct {
	suffix string
	action trade.SettlementInstructionAction
	asset  string
	amount decimal.Decimal
}

func createMatchSettlementRecords(
	ctx context.Context,
	instructionModel models.TTradeSettlementInstructionModel,
	eventModel models.TBizTradeEventModel,
	contractOrderModel models.TTradeOrderContractModel,
	symbol *models.TTradeSymbol,
	order *models.TTradeOrder,
	fill *models.TTradeFill,
	now int64,
) error {
	if symbol == nil || order == nil || fill == nil || fill.Id <= 0 || fill.FillNo == "" || fill.MatchNo == "" {
		return fmt.Errorf("invalid matched fill settlement context")
	}
	specs, err := buildFillSettlementInstructions(ctx, contractOrderModel, symbol, order, fill)
	if err != nil {
		return err
	}
	stepNo := int64(0)
	for _, spec := range specs {
		if !spec.amount.IsPositive() || spec.asset == "" || spec.action == trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_UNKNOWN {
			continue
		}
		stepNo++
		instruction := &models.TTradeSettlementInstruction{
			TenantId:      fill.TenantId,
			InstructionNo: derivedTradeBizNo(fill.FillNo, spec.suffix),
			BizType:       "fill",
			BizId:         fill.FillNo,
			BatchNo:       fill.MatchNo,
			FillId:        fill.Id,
			OrderId:       order.Id,
			ReservationNo: order.OrderNo,
			UserId:        order.UserId,
			Action:        int64(spec.action),
			Asset:         spec.asset,
			Amount:        spec.amount,
			StepNo:        stepNo,
			Status:        int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING),
			NextRetryAt:   now,
			CreateTimes:   now,
			UpdateTimes:   now,
		}
		if err := insertSettlementInstructionIdempotent(ctx, instructionModel, instruction); err != nil {
			return err
		}
	}

	payload, err := json.Marshal(fillEventPayload{
		MatchNo: fill.MatchNo, FillNo: fill.FillNo, OrderNo: order.OrderNo,
		OrderID: order.Id, FillID: fill.Id, UserID: order.UserId, SymbolID: order.SymbolId,
		ProductType: order.ProductType, Side: order.Side, PositionSide: order.PositionSide,
		Price: fill.Price.String(), Qty: fill.Qty.String(), Amount: fill.Amount.String(),
		Fee: fill.Fee.String(), FeeAsset: fill.FeeAsset, OrderStatus: order.Status,
	})
	if err != nil {
		return err
	}
	if err := insertMatchOutboxEvent(ctx, eventModel, order, derivedTradeBizNo(fill.FillNo, "FILL"), "FILL_CREATED", fill.FillNo, "fill", string(payload), now); err != nil {
		return err
	}
	orderEventType := "ORDER_PART_FILLED"
	if order.Status == int64(trade.OrderStatus_ORDER_STATUS_FILLED) {
		orderEventType = "ORDER_FILLED"
	}
	if err := insertMatchOutboxEvent(ctx, eventModel, order, derivedTradeBizNo(fill.FillNo, "ORDER"), orderEventType, order.OrderNo, "order", string(payload), now); err != nil {
		return err
	}
	if order.ProductType == int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		if err := insertMatchOutboxEvent(ctx, eventModel, order, derivedTradeBizNo(fill.FillNo, "POSITION"), "POSITION_FILL_REQUIRED", fill.FillNo, "position_instruction", string(payload), now); err != nil {
			return err
		}
	}
	return nil
}

func derivedTradeBizNo(base, suffix string) string {
	value := base + "-" + suffix
	if len(value) <= 64 {
		return value
	}
	digest := sha256.Sum256([]byte(value))
	return base[:47] + "-" + fmt.Sprintf("%x", digest[:8])
}

func buildFillSettlementInstructions(ctx context.Context, contractOrderModel models.TTradeOrderContractModel, symbol *models.TTradeSymbol, order *models.TTradeOrder, fill *models.TTradeFill) ([]settlementInstructionSpec, error) {
	if order.ProductType == int64(trade.ProductType_PRODUCT_TYPE_SPOT) {
		if order.Side == int64(common.Side_SIDE_BUY) {
			return []settlementInstructionSpec{
				{suffix: "CONSUME", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CONSUME_FROZEN, asset: symbol.QuoteAsset, amount: fill.Amount},
				{suffix: "CREDIT", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE, asset: symbol.BaseAsset, amount: toTradeMinorAmount(fill.Qty)},
				{suffix: "FEE", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE, asset: fill.FeeAsset, amount: fill.Fee},
			}, nil
		}
		return []settlementInstructionSpec{
			{suffix: "CONSUME", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CONSUME_FROZEN, asset: symbol.BaseAsset, amount: toTradeMinorAmount(fill.Qty)},
			{suffix: "CREDIT", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE, asset: symbol.QuoteAsset, amount: fill.Amount},
			{suffix: "FEE", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE, asset: fill.FeeAsset, amount: fill.Fee},
		}, nil
	}
	if order.ProductType != int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		return nil, fmt.Errorf("unsupported matched product type: %d", order.ProductType)
	}
	contractOrder, err := contractOrderModel.FindOneByTenantIdOrderId(ctx, order.TenantId, order.Id)
	if err != nil {
		return nil, err
	}
	specs := make([]settlementInstructionSpec, 0, 2)
	if order.IsReduceOnly != int64(common.YesNo_YES_NO_YES) && contractOrder.MarginAmount.IsPositive() && order.Qty.IsPositive() {
		marginDelta := contractOrder.MarginAmount.Mul(fill.Qty).Div(order.Qty)
		specs = append(specs, settlementInstructionSpec{suffix: "MARGIN", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_ADJUST_MARGIN, asset: contractOrder.MarginAsset, amount: marginDelta})
	}
	if fill.Fee.IsPositive() {
		specs = append(specs, settlementInstructionSpec{suffix: "FEE", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE, asset: fill.FeeAsset, amount: fill.Fee})
	}
	return specs, nil
}

func insertMatchOutboxEvent(ctx context.Context, eventModel models.TBizTradeEventModel, order *models.TTradeOrder, eventNo, eventType, bizID, bizType, payload string, now int64) error {
	exists, err := eventModel.FindOneByTenantIdEventNo(ctx, order.TenantId, eventNo)
	if err == nil {
		if exists.EventType != eventType || exists.BizId != bizID || exists.BizType != bizType {
			return fmt.Errorf("outbox idempotency conflict: %s", eventNo)
		}
		return nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return err
	}
	_, err = eventModel.Insert(ctx, &models.TBizTradeEvent{
		TenantId: order.TenantId, EventNo: eventNo, EventType: eventType,
		BizId: bizID, BizType: bizType, UserId: order.UserId, SymbolId: order.SymbolId,
		ProductType: order.ProductType, OperatorId: 0, Source: int64(trade.SourceType_SOURCE_TYPE_SYSTEM),
		EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20,
		NextRetryAt: now, Payload: payload, CreateTimes: now, UpdateTimes: now,
	})
	return err
}

func insertSettlementInstructionIdempotent(ctx context.Context, model models.TTradeSettlementInstructionModel, instruction *models.TTradeSettlementInstruction) error {
	exists, err := model.FindOneByTenantIdInstructionNo(ctx, instruction.TenantId, instruction.InstructionNo)
	if err == nil {
		if exists.BizId != instruction.BizId || exists.FillId != instruction.FillId || exists.OrderId != instruction.OrderId || exists.Action != instruction.Action || exists.Asset != instruction.Asset || !exists.Amount.Equal(instruction.Amount) {
			return fmt.Errorf("settlement instruction idempotency conflict: %s", instruction.InstructionNo)
		}
		return nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return err
	}
	_, err = model.Insert(ctx, instruction)
	return err
}
