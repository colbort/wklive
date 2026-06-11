package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && symbol.TenantId != tenantId) {
		return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if in.ClientOrderId != "" {
		exists, err := l.svcCtx.TradeOrderModel.FindOneByTenantIdUserIdClientOrderId(l.ctx, tenantId, userId, in.ClientOrderId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if exists != nil {
			return &trade.PlaceOrderResp{Base: helper.OkResp(), Data: orderToProto(exists)}, nil
		}
	}

	orderType := in.OrderType
	triggerKind := in.TriggerKind
	timeInForce := in.TimeInForce

	price := mustParseFloat(in.Price)
	qty := mustParseFloat(in.Qty)
	amount := mustParseFloat(in.Amount)
	triggerPrice := mustParseFloat(in.TriggerPrice)
	orderType, triggerKind = normalizeOrderTypeAndTriggerKind(orderType, triggerKind, price)
	if !isSupportedOrderType(orderType) || !isSupportedTriggerKind(triggerKind) {
		return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if hasNegativeOrderInput(price, qty, amount, triggerPrice) {
		return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if !isValidOrderPrice(orderType, price) || !isValidOrderTimeInForce(orderType, triggerKind, timeInForce) {
		return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	timeInForce = normalizeOrderTimeInForce(orderType, timeInForce)
	if amount == 0 {
		amountPrice, err := l.orderAmountPrice(symbol, orderType, price)
		if err != nil {
			l.Errorf("place order resolve amount price failed, tenantId=%d userId=%d symbolId=%d orderType=%d price=%v triggerPrice=%v err=%v",
				tenantId, userId, in.SymbolId, orderType, price, triggerPrice, err)
			return nil, err
		}
		if amountPrice <= 0 || qty <= 0 {
			return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		amount = tradeMinorAmountAtPrice(amountPrice, qty)
	}

	if qty <= 0 && amount <= 0 {
		return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if isTriggerKind(triggerKind) && triggerPrice <= 0 {
		return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if timeInForce == trade.TimeInForce_TIME_IN_FORCE_POST_ONLY {
		if orderType != trade.OrderType_ORDER_TYPE_LIMIT || price <= 0 {
			return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		wouldTake, err := l.postOnlyWouldTake(tenantId, in.SymbolId, int64(in.MarketType), int64(in.Side), price)
		if err != nil {
			return nil, err
		}
		if wouldTake {
			return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.PostOnlyOrderWouldMatchImmediately, i18n.Translate(i18n.PostOnlyOrderWouldMatchImmediately, l.ctx))}, nil
		}
	}
	leverage := int64(1)
	if in.MarketType != trade.MarketType_MARKET_TYPE_SPOT {
		var ok bool
		leverage, ok, err = ensureConfiguredLeverage(l.ctx, l.svcCtx.SymbolLeverageCfgModel, tenantId, symbol, in.MarginMode, in.Leverage)
		if err != nil {
			return nil, err
		}
		if !ok {
			return &trade.PlaceOrderResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
	}

	orderNo, err := l.svcCtx.GenerateBizNo(l.ctx, "TRD")
	if err != nil {
		return nil, err
	}
	marginAsset := marginAssetForSymbol(symbol)
	now := utils.NowMillis()
	order := &models.TTradeOrder{
		TenantId:      tenantId,
		OrderNo:       orderNo,
		ClientOrderId: in.ClientOrderId,
		UserId:        userId,
		SymbolId:      in.SymbolId,
		MarketType:    int64(in.MarketType),
		Side:          int64(in.Side),
		PositionSide:  int64(in.PositionSide),
		OrderType:     int64(orderType),
		TimeInForce:   int64(timeInForce),
		Status:        int64(trade.OrderStatus_ORDER_STATUS_FREEZING),
		Price:         price,
		Qty:           qty,
		Amount:        amount,
		FilledQty:     0,
		FilledAmount:  0,
		AvgPrice:      0,
		Fee:           0,
		FeeAsset:      marginAsset,
		Source:        int64(in.OrderSource),
		IsReduceOnly:  in.IsReduceOnly,
		IsCloseOnly:   in.IsCloseOnly,
		TriggerPrice:  triggerPrice,
		TriggerType:   int64(in.TriggerType),
		TriggerKind:   int64(triggerKind),
		BizExt:        sql.NullString{String: "", Valid: false},
		CreateTimes:   now,
		UpdateTimes:   now,
	}
	var (
		frozenAsset  string
		frozenAmount float64
		freezeNo     string
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
				TenantId:     tenantId,
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

		frozenAsset, frozenAmount = marginAsset, amount
		contract := &models.TTradeOrderContract{
			TenantId:          tenantId,
			OrderId:           order.Id,
			MarginMode:        int64(in.MarginMode),
			Leverage:          leverage,
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

	freezeNo, err = freezeOrderAsset(l.svcCtx, l.ctx, order, symbol, frozenAsset, frozenAmount)
	if err != nil {
		l.Errorf("place order freeze asset failed, tenantId=%d userId=%d orderNo=%s symbolId=%d marketType=%d frozenAsset=%s frozenAmount=%v err=%v",
			tenantId, userId, order.OrderNo, in.SymbolId, in.MarketType, frozenAsset, frozenAmount, err)
		order.Status = int64(trade.OrderStatus_ORDER_STATUS_REJECTED)
		order.CancelReason = fmt.Sprintf("asset freeze failed: %v", err)
		order.UpdateTimes = utils.NowMillis()
		if updateErr := l.svcCtx.TradeOrderModel.Update(l.ctx, order); updateErr != nil {
			l.Errorf("update rejected order failed, orderNo=%s err=%v", order.OrderNo, updateErr)
		}
		return nil, err
	}
	if freezeNo != "" {
		ext := orderAssetExt{FreezeNo: freezeNo}
		if isTriggerKind(triggerKind) {
			ext.OriginalOrderType = int64(orderType)
			ext.TriggerPrice = fmt.Sprintf("%v", triggerPrice)
		}
		extValue, err := marshalOrderAssetExt(ext)
		if err != nil {
			if compensateErr := unfreezeOrderAsset(l.svcCtx, l.ctx, order, freezeNo, frozenAmount, "trade place order compensate unfreeze"); compensateErr != nil {
				l.Errorf("place order compensate unfreeze failed after marshal ext failed, tenantId=%d userId=%d orderNo=%s freezeNo=%s amount=%v err=%v compensateErr=%v",
					tenantId, userId, order.OrderNo, freezeNo, frozenAmount, err, compensateErr)
				return nil, i18n.StatusError(l.ctx, i18n.InternalServerError)
			}
			l.Errorf("place order marshal asset ext failed after freeze, tenantId=%d userId=%d orderNo=%s freezeNo=%s amount=%v err=%v",
				tenantId, userId, order.OrderNo, freezeNo, frozenAmount, err)
			return nil, err
		}
		order.BizExt = sql.NullString{String: extValue, Valid: extValue != ""}
		order.Status = statusAfterFreeze(triggerKind)
		order.UpdateTimes = utils.NowMillis()
		if err := l.svcCtx.TradeOrderModel.Update(l.ctx, order); err != nil {
			if compensateErr := unfreezeOrderAsset(l.svcCtx, l.ctx, order, freezeNo, frozenAmount, "trade place order compensate unfreeze"); compensateErr != nil {
				l.Errorf("place order compensate unfreeze failed after update order failed, tenantId=%d userId=%d orderNo=%s freezeNo=%s amount=%v err=%v compensateErr=%v",
					tenantId, userId, order.OrderNo, freezeNo, frozenAmount, err, compensateErr)
				return nil, i18n.StatusError(l.ctx, i18n.InternalServerError)
			}
			l.Errorf("place order update order after freeze failed, tenantId=%d userId=%d orderNo=%s freezeNo=%s amount=%v err=%v",
				tenantId, userId, order.OrderNo, freezeNo, frozenAmount, err)
			return nil, err
		}
	} else {
		order.Status = statusAfterFreeze(triggerKind)
		order.UpdateTimes = utils.NowMillis()
		if err := l.svcCtx.TradeOrderModel.Update(l.ctx, order); err != nil {
			return nil, err
		}
	}
	if err := syncOrderBookCache(l.svcCtx, l.ctx, order); err != nil {
		l.Errorf("sync redis order book after place order failed, orderId=%d err=%v", order.Id, err)
	}

	return &trade.PlaceOrderResp{Base: helper.OkResp(), Data: orderToProto(order)}, nil
}

func (l *PlaceOrderLogic) orderAmountPrice(symbol *models.TTradeSymbol, orderType trade.OrderType, price float64) (float64, error) {
	switch {
	case orderType == trade.OrderType_ORDER_TYPE_LIMIT:
		return price, nil
	case orderType == trade.OrderType_ORDER_TYPE_MARKET:
		if symbol == nil {
			return 0, nil
		}
		return l.svcCtx.LastPrice(l.ctx, symbol.Symbol)
	default:
		return 0, nil
	}
}

func (l *PlaceOrderLogic) postOnlyWouldTake(tenantID, symbolID, marketType, side int64, price float64) (bool, error) {
	oppositeSide := int64(common.Side_SIDE_SELL)
	if side == int64(common.Side_SIDE_SELL) {
		oppositeSide = int64(common.Side_SIDE_BUY)
	}
	orders, err := l.svcCtx.TradeOrderModel.FindOpenMatchOrders(
		l.ctx,
		tenantID,
		symbolID,
		marketType,
		oppositeSide,
		matchableOrderStatuses(),
		int64(trade.OrderType_ORDER_TYPE_MARKET),
		1,
	)
	if err != nil || len(orders) == 0 {
		return false, err
	}
	opposite := orders[0]
	if opposite.OrderType == int64(trade.OrderType_ORDER_TYPE_MARKET) {
		return true, nil
	}
	if side == int64(common.Side_SIDE_BUY) {
		return price+orderFillEpsilon >= opposite.Price, nil
	}
	return opposite.Price+orderFillEpsilon >= price, nil
}
