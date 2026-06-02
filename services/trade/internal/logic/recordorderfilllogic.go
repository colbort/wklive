package logic

import (
	"context"
	"errors"
	"math"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	errInvalidOrderFill = errors.New("invalid order fill")
	errOrderNotOpen     = errors.New("order is not open")
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
	now := utils.NowMillis()
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		fillModel := models.NewTTradeFillModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeFillModel)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeOrderModel)
		return recordOrderFillWithModels(ctx, fillModel, orderModel, in.Fill, now)
	})
	if errors.Is(err, errInvalidOrderFill) {
		return &trade.InternalCommonResp{Base: helper.GetErrResp(400, errInvalidOrderFill.Error())}, nil
	}
	if errors.Is(err, models.ErrNotFound) {
		return &trade.InternalCommonResp{Base: helper.GetErrResp(404, "order not found")}, nil
	}
	if errors.Is(err, errOrderNotOpen) {
		return &trade.InternalCommonResp{Base: helper.GetErrResp(400, errOrderNotOpen.Error())}, nil
	}
	if err != nil {
		return nil, err
	}
	return &trade.InternalCommonResp{Base: helper.OkResp()}, nil
}

func recordOrderFillWithModels(ctx context.Context, fillModel models.TradeFillModel, orderModel models.TradeOrderModel, in *trade.TradeFill, now int64) error {
	fill, err := tradeFillFromProto(in, now)
	if err != nil {
		return err
	}
	exists, err := fillModel.FindOneByTenantIdFillNo(ctx, fill.TenantId, fill.FillNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if exists != nil {
		return nil
	}

	order, err := findOrderForFill(ctx, orderModel, in)
	if err != nil {
		return err
	}
	if !isMatchableOrderStatus(order.Status) {
		return errOrderNotOpen
	}
	if !fillMatchesOrder(order, fill) {
		return errInvalidOrderFill
	}
	if !canApplyOrderFill(order, fill) {
		return errInvalidOrderFill
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
	if fill.MarketType <= 0 {
		fill.MarketType = order.MarketType
	}
	if fill.Side <= 0 {
		fill.Side = order.Side
	}
	if fill.PositionSide <= 0 {
		fill.PositionSide = order.PositionSide
	}

	if _, err = fillModel.Insert(ctx, fill); err != nil {
		return err
	}
	applyFillToOrder(order, fill, now)
	return orderModel.Update(ctx, order)
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
	if fill.MarketType > 0 && fill.MarketType != order.MarketType {
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
	if order.Amount > 0 && (fill.Amount <= 0 || fill.Amount > math.Max(order.Amount-order.FilledAmount, 0)+orderMatchAmountEpsilon) {
		return false
	}
	if order.Qty > 0 {
		return fill.Qty > 0 && fill.Qty <= math.Max(order.Qty-order.FilledQty, 0)+orderFillEpsilon
	}
	return order.Amount > 0
}

func tradeFillFromProto(fill *trade.TradeFill, now int64) (*models.TTradeFill, error) {
	if fill == nil || fill.TenantId <= 0 || fill.FillNo == "" || (fill.OrderId <= 0 && fill.OrderNo == "") {
		return nil, errInvalidOrderFill
	}
	price := mustParseFloat(fill.Price)
	qty := mustParseFloat(fill.Qty)
	amount := mustParseFloat(fill.Amount)
	if amount <= 0 && price > 0 && qty > 0 {
		amount = price * qty
	}
	if qty <= 0 && amount <= 0 {
		return nil, errInvalidOrderFill
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
		TenantId:      fill.TenantId,
		FillNo:        fill.FillNo,
		OrderId:       fill.OrderId,
		OrderNo:       fill.OrderNo,
		UserId:        fill.UserId,
		SymbolId:      fill.SymbolId,
		MarketType:    int64(fill.MarketType),
		Side:          int64(fill.Side),
		PositionSide:  int64(fill.PositionSide),
		Price:         price,
		Qty:           qty,
		Amount:        amount,
		Fee:           mustParseFloat(fill.Fee),
		FeeAsset:      fill.FeeAsset,
		LiquidityType: int64(fill.LiquidityType),
		RealizedPnl:   mustParseFloat(fill.RealizedPnl),
		MatchTime:     matchTime,
		CreateTimes:   createTimes,
	}, nil
}

func findOrderForFill(ctx context.Context, orderModel models.TradeOrderModel, fill *trade.TradeFill) (*models.TTradeOrder, error) {
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
	order.FilledQty += fill.Qty
	order.FilledAmount += fill.Amount
	if order.Qty > 0 && order.FilledQty > order.Qty && order.FilledQty-order.Qty <= orderFillEpsilon {
		order.FilledQty = order.Qty
	}
	if order.Amount > 0 && order.FilledAmount > order.Amount && order.FilledAmount-order.Amount <= orderFillEpsilon {
		order.FilledAmount = order.Amount
	}
	if order.FilledQty > 0 && order.FilledAmount > 0 {
		order.AvgPrice = order.FilledAmount / order.FilledQty
	}
	order.Fee += fill.Fee
	if fill.FeeAsset != "" {
		order.FeeAsset = fill.FeeAsset
	}
	order.Status = orderStatusAfterFill(order)
	order.UpdateTimes = now
}
