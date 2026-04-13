package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type PlaceOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceOrderLogic {
	return &PlaceOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 下单
func (l *PlaceOrderLogic) PlaceOrder(in *trade.PlaceOrderReq) (*trade.PlaceOrderResp, error) {
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && symbol.TenantId != in.TenantId) {
		return &trade.PlaceOrderResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if in.ClientOrderId != "" {
		exists, err := l.svcCtx.TradeOrderModel.FindOneByTenantIdUserIdClientOrderId(l.ctx, in.TenantId, in.UserId, in.ClientOrderId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if exists != nil {
			return &trade.PlaceOrderResp{Base: helper.OkResp(), Order: orderToProto(exists)}, nil
		}
	}

	price := mustParseFloat(in.Price)
	qty := mustParseFloat(in.Qty)
	amount := mustParseFloat(in.Amount)
	if amount == 0 {
		if price > 0 && qty > 0 {
			amount = price * qty
		} else {
			amount = qty
		}
	}
	if qty <= 0 && amount <= 0 {
		return &trade.PlaceOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	orderNo, err := l.svcCtx.GenerateBizNo(l.ctx, "TRD")
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	order := &models.TTradeOrder{
		TenantId:      in.TenantId,
		OrderNo:       orderNo,
		ClientOrderId: in.ClientOrderId,
		UserId:        in.UserId,
		SymbolId:      in.SymbolId,
		MarketType:    int64(in.MarketType),
		Side:          int64(in.Side),
		PositionSide:  int64(in.PositionSide),
		OrderType:     int64(in.OrderType),
		TimeInForce:   int64(in.TimeInForce),
		Status:        int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		Price:         price,
		Qty:           qty,
		Amount:        amount,
		FilledQty:     0,
		FilledAmount:  0,
		AvgPrice:      0,
		Fee:           0,
		FeeAsset:      marginAssetForSymbol(symbol),
		Source:        int64(in.OrderSource),
		IsReduceOnly:  in.IsReduceOnly,
		IsCloseOnly:   in.IsCloseOnly,
		TriggerPrice:  mustParseFloat(in.TriggerPrice),
		TriggerType:   int64(in.TriggerType),
		BizExt:        sql.NullString{String: "", Valid: false},
		CreateTimes:   now,
		UpdateTimes:   now,
	}
	var (
		frozenAsset  string
		frozenAmount float64
	)
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeOrderModel)
		spotModel := models.NewTTradeOrderSpotModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeOrderSpotModel)
		contractModel := models.NewTTradeOrderContractModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeOrderContractModel)

		res, err := orderModel.Insert(ctx, order)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		order.Id = id

		if in.MarketType == trade.MarketType_MARKET_TYPE_SPOT {
			frozenAsset, frozenAmount = spotFrozenAssetAndAmount(symbol, in.Side, qty, amount)
			spot := &models.TTradeOrderSpot{
				TenantId:     in.TenantId,
				OrderId:      order.Id,
				FrozenAsset:  frozenAsset,
				FrozenAmount: frozenAmount,
				SettleAsset:  symbol.SettleAsset,
				SettleAmount: amount,
				CreateTimes:  now,
				UpdateTimes:  now,
			}
			if _, err = spotModel.Insert(ctx, spot); err != nil {
				return err
			}
			return nil
		}

		marginAsset := marginAssetForSymbol(symbol)
		frozenAsset, frozenAmount = marginAsset, amount
		contract := &models.TTradeOrderContract{
			TenantId:          in.TenantId,
			OrderId:           order.Id,
			MarginMode:        int64(in.MarginMode),
			Leverage:          ensureLeverage(symbol, in.Leverage),
			MarginAsset:       marginAsset,
			MarginAmount:      amount,
			ClosePositionType: 0,
			LiquidationPrice:  0,
			TakeProfitPrice:   mustParseFloat(in.TakeProfitPrice),
			StopLossPrice:     mustParseFloat(in.StopLossPrice),
			CreateTimes:       now,
			UpdateTimes:       now,
		}
		if _, err = contractModel.Insert(ctx, contract); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if _, err = freezeOrderAsset(l.svcCtx, l.ctx, order, symbol, frozenAsset, frozenAmount); err != nil {
		order.Status = int64(trade.OrderStatus_ORDER_STATUS_REJECTED)
		order.CancelReason = fmt.Sprintf("asset freeze failed: %v", err)
		order.UpdateTimes = utils.NowMillis()
		if updateErr := l.svcCtx.TradeOrderModel.Update(l.ctx, order); updateErr != nil {
			l.Errorf("update rejected order failed, orderNo=%s err=%v", order.OrderNo, updateErr)
		}
		return nil, err
	}

	return &trade.PlaceOrderResp{Base: helper.OkResp(), Order: orderToProto(order)}, nil
}
