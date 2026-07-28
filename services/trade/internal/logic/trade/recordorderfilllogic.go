package tradelogic

import (
	"context"
	"errors"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RecordOrderFillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordOrderFillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordOrderFillLogic {
	return &RecordOrderFillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 记录订单成交信息
func (l *RecordOrderFillLogic) RecordOrderFill(in *trade.RecordOrderFillReq) (*trade.InternalCommonResp, error) {
	if in.Fill == nil {
		return &trade.InternalCommonResp{Base: helper.OkResp()}, nil
	}
	if in.Fill.MatchNo == "" {
		matchNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "MAT", "")
		if err != nil {
			return nil, err
		}
		in.Fill.MatchNo = matchNo
	}
	now := utils.NowMillis()
	err := helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		fillModel := models.NewTTradeFillModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis)
		contractOrderModel := models.NewTTradeOrderContractModel(conn, l.svcCtx.Config.CacheRedis)
		fill, order, err := recordOrderFillWithModels(ctx, fillModel, orderModel, in.Fill, now)
		if err != nil || order == nil {
			return err
		}
		symbol, err := l.svcCtx.TradeSymbolModel.FindOne(ctx, order.SymbolId)
		if err != nil {
			return err
		}
		return createMatchSettlementRecords(ctx, instructionModel, eventModel, contractOrderModel, symbol, order, fill, now)
	})
	if i18n.IsStatusError(err, i18n.ParamError) {
		return &trade.InternalCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if errors.Is(err, models.ErrNotFound) {
		return &trade.InternalCommonResp{Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if i18n.IsStatusError(err, i18n.OperationNotAllowed) {
		return &trade.InternalCommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	return &trade.InternalCommonResp{Base: helper.OkResp()}, nil
}

func recordOrderFillWithModels(ctx context.Context, fillModel models.TTradeFillModel, orderModel models.TTradeOrderModel, in *trade.TradeFill, now int64) (*models.TTradeFill, *models.TTradeOrder, error) {
	fill, err := tradeFillFromProto(in, now)
	if err != nil {
		return nil, nil, err
	}
	exists, err := fillModel.FindOneByTenantIdFillNo(ctx, fill.TenantId, fill.FillNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, nil, err
	}
	if exists != nil {
		order, findErr := findOrderForFill(ctx, orderModel, in)
		return exists, order, findErr
	}

	order, err := findOrderForFill(ctx, orderModel, in)
	if err != nil {
		return nil, nil, err
	}
	if !helpers.IsMatchableOrderStatus(order.Status) {
		return nil, nil, i18n.StatusError(ctx, i18n.OperationNotAllowed)
	}
	if !fillMatchesOrder(order, fill) {
		return nil, nil, i18n.StatusError(ctx, i18n.ParamError)
	}
	if !canApplyOrderFill(order, fill) {
		return nil, nil, i18n.StatusError(ctx, i18n.ParamError)
	}

	if fill.OrderId <= 0 {
		fill.OrderId = order.Id
	}
	if fill.OrderNo == "" {
		fill.OrderNo = order.OrderNo
	}
	if fill.UserId <= 0 {
		fill.UserId = order.UserId
	}
	if fill.SymbolId <= 0 {
		fill.SymbolId = order.SymbolId
	}
	if fill.ProductType <= 0 {
		fill.ProductType = order.ProductType
	}
	if fill.ContractType <= 0 {
		fill.ContractType = order.ContractType
	}
	if fill.ContractValueType <= 0 {
		fill.ContractValueType = order.ContractValueType
	}
	if fill.Side <= 0 {
		fill.Side = order.Side
	}
	if fill.PositionSide <= 0 {
		fill.PositionSide = order.PositionSide
	}

	result, err := fillModel.Insert(ctx, fill)
	if err != nil {
		return nil, nil, err
	}
	fill.Id, _ = result.LastInsertId()
	applyFillToOrder(order, fill, now)
	if err := orderModel.Update(ctx, order); err != nil {
		return nil, nil, err
	}
	return fill, order, nil
}

func fillMatchesOrder(order *models.TTradeOrder, fill *models.TTradeFill) bool {
	if order == nil || fill == nil {
		return false
	}
	if fill.OrderNo != "" && fill.OrderNo != order.OrderNo {
		return false
	}
	if fill.UserId > 0 && fill.UserId != order.UserId {
		return false
	}
	if fill.SymbolId > 0 && fill.SymbolId != order.SymbolId {
		return false
	}
	if fill.ProductType > 0 && fill.ProductType != order.ProductType {
		return false
	}
	if fill.Side > 0 && fill.Side != order.Side {
		return false
	}
	if fill.PositionSide > 0 && fill.PositionSide != order.PositionSide {
		return false
	}
	return true
}

func canApplyOrderFill(order *models.TTradeOrder, fill *models.TTradeFill) bool {
	if order == nil || fill == nil {
		return false
	}
	if !canApplyOrderFillPrice(order, fill) {
		return false
	}
	if order.Qty.IsPositive() {
		return fill.Qty.IsPositive() && fill.Qty.LessThanOrEqual(decimal.Max(order.Qty.Sub(order.FilledQty), decimal.Zero))
	}
	return order.Amount.IsPositive() && fill.Amount.IsPositive() &&
		fill.Amount.LessThanOrEqual(decimal.Max(order.Amount.Sub(order.FilledAmount), decimal.Zero))
}

func canApplyOrderFillPrice(order *models.TTradeOrder, fill *models.TTradeFill) bool {
	if order == nil || fill == nil || order.OrderType != int64(trade.OrderType_ORDER_TYPE_LIMIT) || !order.Price.IsPositive() || !fill.Price.IsPositive() {
		return true
	}
	if order.Side == int64(common.Side_SIDE_BUY) {
		return fill.Price.LessThanOrEqual(order.Price)
	}
	if order.Side == int64(common.Side_SIDE_SELL) {
		return fill.Price.GreaterThanOrEqual(order.Price)
	}
	return false
}

func tradeFillFromProto(fill *trade.TradeFill, now int64) (*models.TTradeFill, error) {
	if fill == nil || fill.TenantId <= 0 || fill.FillNo == "" || fill.MatchNo == "" || (fill.OrderId <= 0 && fill.OrderNo == "") {
		return nil, i18n.StatusError(context.Background(), i18n.ParamError)
	}
	price := helpers.MustParseFloat(fill.Price)
	qty := helpers.MustParseFloat(fill.Qty)
	amount := helpers.MustParseFloat(fill.Amount)
	if !amount.IsPositive() && price.IsPositive() && qty.IsPositive() {
		amount = helpers.TradeMinorAmountAtPrice(price, qty)
	}
	if !price.IsPositive() || !qty.IsPositive() || !amount.IsPositive() {
		return nil, i18n.StatusError(context.Background(), i18n.ParamError)
	}
	createTimes := fill.CreateTimes
	if createTimes <= 0 {
		createTimes = now
	}
	matchTime := fill.MatchTime
	if matchTime <= 0 {
		matchTime = createTimes
	}
	return &models.TTradeFill{
		TenantId:          fill.TenantId,
		FillNo:            fill.FillNo,
		MatchNo:           fill.MatchNo,
		OrderId:           fill.OrderId,
		OrderNo:           fill.OrderNo,
		UserId:            fill.UserId,
		SymbolId:          fill.SymbolId,
		ProductType:       int64(fill.ProductType),
		ContractType:      int64(fill.ContractType),
		ContractValueType: int64(fill.ContractValueType),
		Side:              int64(fill.Side),
		PositionSide:      int64(fill.PositionSide),
		Price:             price,
		Qty:               qty,
		Amount:            amount,
		Fee:               helpers.MustParseFloat(fill.Fee),
		FeeAsset:          fill.FeeAsset,
		LiquidityType:     int64(fill.LiquidityType),
		RealizedPnl:       helpers.MustParseFloat(fill.RealizedPnl),
		SettlementStatus:  int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_PENDING),
		MatchTime:         matchTime,
		CreateTimes:       createTimes,
	}, nil
}

func findOrderForFill(ctx context.Context, orderModel models.TTradeOrderModel, fill *trade.TradeFill) (*models.TTradeOrder, error) {
	if fill == nil {
		return nil, models.ErrNotFound
	}
	if fill.OrderId > 0 {
		order, err := orderModel.FindOneForUpdate(ctx, fill.OrderId)
		if err != nil {
			return nil, err
		}
		if order.TenantId != fill.TenantId {
			return nil, models.ErrNotFound
		}
		return order, nil
	}
	if fill.OrderNo != "" {
		return orderModel.FindOneByTenantIdOrderNoForUpdate(ctx, fill.TenantId, fill.OrderNo)
	}
	return nil, models.ErrNotFound
}

func applyFillToOrder(order *models.TTradeOrder, fill *models.TTradeFill, now int64) {
	previousFilledQty := order.FilledQty
	order.FilledQty = order.FilledQty.Add(fill.Qty)
	order.FilledAmount = order.FilledAmount.Add(fill.Amount)
	if order.Qty.IsPositive() && order.FilledQty.GreaterThan(order.Qty) {
		order.FilledQty = order.Qty
	}
	if order.ProductType == int64(common.ProductType_PRODUCT_TYPE_SPOT) && order.Amount.IsPositive() && order.FilledAmount.GreaterThan(order.Amount) {
		order.FilledAmount = order.Amount
	}
	if order.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) && fill.Price.IsPositive() {
		order.AvgPrice = contractAveragePrice(order.AvgPrice, previousFilledQty, fill.Price, fill.Qty, order.ContractValueType)
	} else if order.FilledQty.IsPositive() && order.FilledAmount.IsPositive() {
		order.AvgPrice = helpers.FromTradeMinorAmount(order.FilledAmount).Div(order.FilledQty)
	}
	order.Fee = order.Fee.Add(fill.Fee)
	if fill.FeeAsset != "" {
		order.FeeAsset = fill.FeeAsset
	}
	order.Status = helpers.OrderStatusAfterFill(order)
	order.Version++
	order.UpdateTimes = now
}
