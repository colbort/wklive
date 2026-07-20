package logic

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

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
	"google.golang.org/protobuf/proto"
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
		return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	configTenantId := symbol.TenantId
	requestBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(in)
	if err != nil {
		return nil, err
	}
	requestDigest := sha256.Sum256(requestBytes)
	requestHash := hex.EncodeToString(requestDigest[:])
	if in.ClientOrderId != "" {
		exists, err := l.svcCtx.TradeOrderModel.FindOneByTenantIdUserIdProductTypeClientOrderId(l.ctx, tenantId, userId, symbol.ProductType, sql.NullString{String: in.ClientOrderId, Valid: true})
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if exists != nil {
			if exists.RequestHash != "" && exists.RequestHash != requestHash {
				return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, "client_order_id already exists with different order parameters")}, nil
			}
			return &trade.PlaceOrderResp{Base: helper.OkResp(), Data: orderToProto(exists)}, nil
		}
	}

	orderType := in.OrderType
	triggerKind := in.TriggerKind
	timeInForce := in.TimeInForce
	isSeconds := symbol.ProductType == int64(trade.ProductType_PRODUCT_TYPE_SECONDS)
	var secondsCfg *models.TTradeSymbolSeconds

	price := mustParseFloat(in.Price)
	qty := mustParseFloat(in.Qty)
	amount := mustParseFloat(in.Amount)
	triggerPrice := mustParseFloat(in.TriggerPrice)
	if isSeconds {
		orderType, triggerKind, timeInForce = trade.OrderType_ORDER_TYPE_UNKNOWN, trade.TriggerKind_TRIGGER_KIND_NONE, trade.TimeInForce_TIME_IN_FORCE_UNKNOWN
		if in.SecondsDirection < 1 || in.SecondsDirection > 2 || in.DurationSeconds <= 0 || !amount.IsPositive() {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		secondsCfg, err = l.svcCtx.TradeSymbolSecondsModel.FindOneByTenantIdSymbolIdDurationSeconds(l.ctx, configTenantId, symbol.Id, in.DurationSeconds)
		if errors.Is(err, models.ErrNotFound) {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
		}
		if err != nil {
			return nil, err
		}
		if (in.SecondsDirection == 1 && secondsCfg.UpEnabled != 1) || (in.SecondsDirection == 2 && secondsCfg.DownEnabled != 1) || amount.LessThan(secondsCfg.MinStake) || (secondsCfg.MaxStake.IsPositive() && amount.GreaterThan(secondsCfg.MaxStake)) {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
		}
	} else {
		orderType, triggerKind = normalizeOrderTypeAndTriggerKind(orderType, triggerKind, price)
		if !isSupportedOrderType(orderType) || !isSupportedTriggerKind(triggerKind) {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
	}
	if hasNegativeOrderInput(price, qty, amount, triggerPrice) {
		return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if !isSeconds && (!isValidOrderPrice(orderType, price) || !isValidOrderTimeInForce(orderType, triggerKind, timeInForce)) {
		return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if !isSeconds {
		timeInForce = normalizeOrderTimeInForce(orderType, timeInForce)
	}
	if !isSeconds && amount.IsZero() {
		amountPrice, err := l.orderAmountPrice(symbol, orderType, price)
		if err != nil {
			l.Errorf("place order resolve amount price failed, tenantId=%d userId=%d symbolId=%d orderType=%d price=%v triggerPrice=%v err=%v",
				tenantId, userId, in.SymbolId, orderType, price, triggerPrice, err)
			return nil, err
		}
		if !amountPrice.IsPositive() || !qty.IsPositive() {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		amount = tradeMinorAmountAtPrice(amountPrice, qty)
	}

	if !qty.IsPositive() && !amount.IsPositive() {
		return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if !isSeconds && isTriggerKind(triggerKind) && !triggerPrice.IsPositive() {
		return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if timeInForce == trade.TimeInForce_TIME_IN_FORCE_POST_ONLY {
		if orderType != trade.OrderType_ORDER_TYPE_LIMIT || !price.IsPositive() {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		wouldTake, err := l.postOnlyWouldTake(tenantId, in.SymbolId, symbol.ProductType, int64(in.Side), price)
		if err != nil {
			return nil, err
		}
		if wouldTake {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.PostOnlyOrderWouldMatchImmediately, i18n.Translate(i18n.PostOnlyOrderWouldMatchImmediately, l.ctx))}, nil
		}
	}
	leverage := int64(1)
	if isDerivativeProduct(trade.ProductType(symbol.ProductType)) {
		var ok bool
		leverage, ok, err = ensureConfiguredLeverage(l.ctx, l.svcCtx.SymbolLeverageCfgModel, l.svcCtx.SymbolLeverageDefaultModel, configTenantId, symbol, in.MarginMode, in.Leverage)
		if err != nil {
			return nil, err
		}
		if !ok {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
	}

	orderNo, err := l.svcCtx.GenerateBizNo(l.ctx, "TRD")
	if err != nil {
		return nil, err
	}
	marginAsset := marginAssetForSymbol(symbol)
	now := utils.NowMillis()
	order := &models.TTradeOrder{
		TenantId:          tenantId,
		OrderNo:           orderNo,
		ClientOrderId:     sql.NullString{String: in.ClientOrderId, Valid: in.ClientOrderId != ""},
		RequestHash:       requestHash,
		UserId:            userId,
		SymbolId:          in.SymbolId,
		ProductType:       symbol.ProductType,
		ContractType:      symbol.ContractType,
		ContractValueType: symbol.ContractValueType,
		Side:              int64(in.Side),
		PositionSide:      int64(in.PositionSide),
		OrderType:         int64(orderType),
		TimeInForce:       int64(timeInForce),
		Status:            int64(trade.OrderStatus_ORDER_STATUS_FREEZING),
		Price:             price,
		Qty:               qty,
		Amount:            amount,
		FilledQty:         decimal.Zero,
		FilledAmount:      decimal.Zero,
		AvgPrice:          decimal.Zero,
		Fee:               decimal.Zero,
		FeeAsset:          marginAsset,
		Source:            int64(in.OrderSource),
		IsReduceOnly:      yesNoToModel(common.YesNo(in.IsReduceOnly), int64(common.YesNo_YES_NO_NO)),
		TriggerPrice:      triggerPrice,
		TriggerType:       int64(in.TriggerType),
		TriggerKind:       int64(triggerKind),
		BizExt:            sql.NullString{String: "", Valid: false},
		CreateTimes:       now,
		UpdateTimes:       now,
	}
	if isSeconds {
		order.Side, order.PositionSide, order.OrderType, order.TimeInForce, order.Price, order.Qty = 0, 0, 0, 0, decimal.Zero, decimal.Zero
	}
	var (
		frozenAsset  string
		frozenAmount decimal.Decimal
		freezeNo     string
	)
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		spotModel := models.NewTTradeOrderSpotModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTTradeOrderContractModel(conn, l.svcCtx.Config.CacheRedis)
		secondsModel := models.NewTTradeOrderSecondsModel(conn, l.svcCtx.Config.CacheRedis)

		res, err := orderModel.Insert(ctx, order)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		order.Id = id

		if symbol.ProductType == int64(trade.ProductType_PRODUCT_TYPE_SPOT) {
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
		if isSeconds {
			frozenAsset, frozenAmount = symbol.SettleAsset, amount
			_, err = secondsModel.Insert(ctx, &models.TTradeOrderSeconds{TenantId: tenantId, OrderId: order.Id, Direction: int64(in.SecondsDirection), DurationSeconds: in.DurationSeconds, StakeAsset: frozenAsset, StakeAmount: amount, PayoutRate: secondsCfg.PayoutRate, SettlementStatus: 0, CreateTimes: now, UpdateTimes: now})
			return err
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
			LiquidationPrice:  decimal.Zero,
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
		l.Errorf("place order freeze asset failed, tenantId=%d userId=%d orderNo=%s symbolId=%d productType=%d frozenAsset=%s frozenAmount=%v err=%v",
			tenantId, userId, order.OrderNo, in.SymbolId, symbol.ProductType, frozenAsset, frozenAmount, err)
		order.Status = int64(trade.OrderStatus_ORDER_STATUS_REJECTED)
		order.CancelReason = fmt.Sprintf("asset freeze failed: %v", err)
		order.UpdateTimes = utils.NowMillis()
		if updateErr := l.svcCtx.TradeOrderModel.Update(l.ctx, order); updateErr != nil {
			l.Errorf("update rejected order failed, orderNo=%s err=%v", order.OrderNo, updateErr)
		}
		return nil, err
	}
	if freezeNo != "" {
		if isSeconds {
			secondsOrder, findErr := l.svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(l.ctx, tenantId, order.Id)
			if findErr != nil {
				return nil, findErr
			}
			secondsOrder.ReservationNo, secondsOrder.UpdateTimes = freezeNo, utils.NowMillis()
			if updateErr := l.svcCtx.TradeOrderSecondsModel.Update(l.ctx, secondsOrder); updateErr != nil {
				return nil, updateErr
			}
		}
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

func (l *PlaceOrderLogic) orderAmountPrice(symbol *models.TTradeSymbol, orderType trade.OrderType, price decimal.Decimal) (decimal.Decimal, error) {
	switch {
	case orderType == trade.OrderType_ORDER_TYPE_LIMIT:
		return price, nil
	case orderType == trade.OrderType_ORDER_TYPE_MARKET:
		return decimal.Zero, nil
	default:
		return decimal.Zero, nil
	}
}

func (l *PlaceOrderLogic) postOnlyWouldTake(tenantID, symbolID, marketType, side int64, price decimal.Decimal) (bool, error) {
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
		return price.GreaterThanOrEqual(opposite.Price), nil
	}
	return opposite.Price.GreaterThanOrEqual(price), nil
}
