package tasklogic

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/domain/contractmath"
	"wklive/services/trade/internal/realtime"
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
	isSeconds := symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_SECONDS)
	var secondsCfg *models.TTradeSymbolSeconds

	price, priceErr := conv.ParseDecimalField(in.Price)
	qty, qtyErr := conv.ParseDecimalField(in.Qty)
	amount, amountErr := conv.ParseDecimalField(in.Amount)
	triggerPrice, triggerPriceErr := conv.ParseDecimalField(in.TriggerPrice)
	if priceErr != nil || qtyErr != nil || amountErr != nil || triggerPriceErr != nil {
		return nil, fmt.Errorf("invalid decimal order input")
	}
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
		if secondsCfg.MaxExposureAmount.IsPositive() {
			lockKey := fmt.Sprintf("trade:seconds:exposure:%d:%d", tenantId, in.SymbolId)
			lockValue := fmt.Sprintf("%d:%d:%d", userId, utils.NowMillis(), in.DurationSeconds)
			if lockErr := acquireTaskLock(l.ctx, l.svcCtx.Redis, lockKey, lockValue); lockErr != nil {
				return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "seconds contract exposure is being updated; retry")}, nil
			}
			defer func() { _ = releaseTaskLock(context.Background(), l.svcCtx.Redis, lockKey, lockValue) }()
			exposure, exposureErr := l.svcCtx.TradeOrderSecondsModel.SumExposure(l.ctx, tenantId, in.SymbolId, []int64{
				int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVATING),
				int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVE),
				int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING),
			})
			if exposureErr != nil {
				return nil, exposureErr
			}
			if exposure.Add(amount).GreaterThan(secondsCfg.MaxExposureAmount) {
				return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "seconds contract exposure limit exceeded")}, nil
			}
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
	if !isSeconds && amount.IsZero() && orderType == trade.OrderType_ORDER_TYPE_LIMIT {
		if !price.IsPositive() || !qty.IsPositive() {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		amount = tradeMinorAmountAtPrice(price, qty)
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
	if isDerivativeProduct(common.ProductType(symbol.ProductType)) {
		var ok bool
		leverage, ok, err = ensureConfiguredLeverage(l.ctx, l.svcCtx.SymbolLeverageCfgModel, l.svcCtx.SymbolLeverageDefaultModel, configTenantId, symbol, in.MarginMode, in.Leverage)
		if err != nil {
			return nil, err
		}
		if !ok {
			return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
	}
	plan, err := l.preparePlaceOrder(tenantId, userId, symbol, secondsCfg, in, orderType, triggerKind, timeInForce, leverage, price, qty, amount)
	if err != nil {
		return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, err.Error())}, nil
	}
	price, qty, amount = plan.price, plan.qty, plan.notional

	orderNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "TRD", "")
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
		IsClosePosition:   int64(common.YesNo_YES_NO_NO),
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
		reservationModel := models.NewTTradeAssetReservationModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)

		res, err := orderModel.Insert(ctx, order)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		order.Id = id
		if plan.frozenAmount.IsPositive() {
			if _, err = reservationModel.Insert(ctx, &models.TTradeAssetReservation{TenantId: tenantId, OrderId: order.Id, ReservationNo: order.OrderNo, Asset: plan.frozenAsset, ReservedAmount: plan.frozenAmount, Status: 1, NextRetryAt: now, CreateTimes: now, UpdateTimes: now}); err != nil {
				return err
			}
		}

		if symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_SPOT) {
			frozenAsset, frozenAmount = plan.frozenAsset, plan.frozenAmount
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
			frozenAsset, frozenAmount = plan.frozenAsset, plan.frozenAmount
			_, err = secondsModel.Insert(ctx, &models.TTradeOrderSeconds{TenantId: tenantId, OrderId: order.Id, Direction: int64(in.SecondsDirection), DurationSeconds: in.DurationSeconds, StakeAsset: frozenAsset, StakeAmount: amount, PayoutRate: secondsCfg.PayoutRate, FeeRate: secondsCfg.FeeRate, StartPriceSource: secondsCfg.StartPriceSource, SettlementPriceSource: secondsCfg.SettlementPriceSource, PriceAlgorithm: secondsCfg.SettlementPriceAlgorithm, SettlementStatus: 0, CreateTimes: now, UpdateTimes: now})
			return err
		}

		frozenAsset, frozenAmount = plan.frozenAsset, plan.frozenAmount
		if plan.reservedCloseQty.IsPositive() {
			if err = positionModel.ReserveCloseQty(ctx, plan.positionID, plan.positionVersion, plan.reservedCloseQty, now); err != nil {
				return err
			}
		}
		contract := &models.TTradeOrderContract{
			TenantId:          tenantId,
			OrderId:           order.Id,
			MarginMode:        int64(in.MarginMode),
			Leverage:          leverage,
			MarginAsset:       marginAsset,
			MarginAmount:      plan.marginAmount,
			ReservedCloseQty:  plan.reservedCloseQty,
			RiskPrice:         plan.riskPrice,
			RiskTierId:        plan.riskTierID,
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
		if in.ClientOrderId != "" {
			exists, findErr := l.svcCtx.TradeOrderModel.FindOneByTenantIdUserIdProductTypeClientOrderId(l.ctx, tenantId, userId, symbol.ProductType, sql.NullString{String: in.ClientOrderId, Valid: true})
			if findErr == nil {
				if exists.RequestHash != "" && exists.RequestHash != requestHash {
					return &trade.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, "client_order_id already exists with different order parameters")}, nil
				}
				return &trade.PlaceOrderResp{Base: helper.OkResp(), Data: orderToProto(exists)}, nil
			}
		}
		return nil, err
	}

	freezeNo, err = freezeOrderAsset(l.svcCtx, l.ctx, order, symbol, frozenAsset, frozenAmount)
	if err != nil {
		l.Errorf("place order freeze asset failed, tenantId=%d userId=%d orderNo=%s symbolId=%d productType=%d frozenAsset=%s frozenAmount=%v err=%v",
			tenantId, userId, order.OrderNo, in.SymbolId, symbol.ProductType, frozenAsset, frozenAmount, err)
		if isDefinitiveAssetFreezeError(err) {
			if rejectErr := l.rejectOrderAfterFreezeFailure(order, plan, err); rejectErr != nil {
				return nil, rejectErr
			}
			return &trade.PlaceOrderResp{Base: helper.OkResp(), Data: orderToProto(order)}, nil
		}
		// Timeout/transport failure does not prove that Asset failed. Keep the
		// order and reservation in FREEZING for idempotent reconciliation.
		if updateErr := l.markAssetReservationRetry(order, err); updateErr != nil {
			l.Errorf("mark uncertain asset reservation failed, orderNo=%s err=%v", order.OrderNo, updateErr)
		}
		return &trade.PlaceOrderResp{Base: helper.OkResp(), Data: orderToProto(order)}, nil
	}
	if err = l.finalizeAcceptedOrder(order, freezeNo, frozenAmount, triggerKind, orderType, triggerPrice, isSeconds); err != nil {
		// Asset has already accepted the idempotent freeze. Never unfreeze here:
		// leave the local order in FREEZING and let reconciliation finalize it.
		_ = l.markAssetReservationRetry(order, err)
		return nil, err
	}
	if err := syncOrderBookCache(l.svcCtx, l.ctx, order); err != nil {
		l.Errorf("sync redis order book after place order failed, orderId=%d err=%v", order.Id, err)
	}
	if !isSeconds && order.Status == int64(trade.OrderStatus_ORDER_STATUS_PENDING) {
		event := realtime.Event{EventNo: derivedTradeBizNo(order.OrderNo, "ACCEPTED"), Type: realtime.EventOrderAccepted, TenantID: order.TenantId, BizID: order.OrderNo, OrderID: order.Id}
		if err := publishTradeOutboxEvent(l.ctx, l.svcCtx, event); err != nil {
			// The event remains pending in the durable outbox and can be retried by
			// ProcessTradeEvents. A broker failure must not roll back the order.
			l.Errorf("publish real-time order accepted event failed, orderId=%d eventNo=%s err=%v", order.Id, event.EventNo, err)
		}
	}

	return &trade.PlaceOrderResp{Base: helper.OkResp(), Data: orderToProto(order)}, nil
}

func (l *PlaceOrderLogic) finalizeAcceptedOrder(order *models.TTradeOrder, freezeNo string, frozenAmount decimal.Decimal, triggerKind trade.TriggerKind, orderType trade.OrderType, triggerPrice decimal.Decimal, isSeconds bool) error {
	ext := orderAssetExt{FreezeNo: freezeNo}
	if isTriggerKind(triggerKind) {
		ext.OriginalOrderType = int64(orderType)
		ext.TriggerPrice = triggerPrice.String()
	}
	extValue, err := marshalOrderAssetExt(ext)
	if err != nil {
		return err
	}
	now := utils.NowMillis()
	if err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		reservationModel := models.NewTTradeAssetReservationModel(conn, l.svcCtx.Config.CacheRedis)
		secondsModel := models.NewTTradeOrderSecondsModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis)

		order.BizExt = sql.NullString{String: extValue, Valid: extValue != ""}
		order.Status = statusAfterFreeze(triggerKind)
		order.UpdateTimes = now
		if err := orderModel.Update(ctx, order); err != nil {
			return err
		}
		if frozenAmount.IsPositive() {
			reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, order.TenantId, order.OrderNo)
			if err != nil {
				return err
			}
			reservation.Status = int64(trade.AssetReservationStatus_ASSET_RESERVATION_STATUS_FROZEN)
			reservation.NextRetryAt = 0
			reservation.LastErrorMsg = ""
			reservation.UpdateTimes = now
			if err := reservationModel.Update(ctx, reservation); err != nil {
				return err
			}
		}
		if isSeconds {
			secondsOrder, err := secondsModel.FindOneByTenantIdOrderId(ctx, order.TenantId, order.Id)
			if err != nil {
				return err
			}
			secondsOrder.ReservationNo = freezeNo
			secondsOrder.FrozenAt = now
			secondsOrder.SettlementStatus = 1
			secondsOrder.UpdateTimes = 0
			if err := secondsModel.Update(ctx, secondsOrder); err != nil {
				return err
			}
		}
		_, err := eventModel.Insert(ctx, &models.TBizTradeEvent{TenantId: order.TenantId, EventNo: derivedTradeBizNo(order.OrderNo, "ACCEPTED"), EventType: realtime.EventOrderAccepted, BizId: order.OrderNo, BizType: "order", UserId: order.UserId, SymbolId: order.SymbolId, ProductType: order.ProductType, OperatorId: order.UserId, Source: int64(trade.SourceType_SOURCE_TYPE_USER), Consumer: tradeEventConsumer(realtime.EventOrderAccepted), EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20, NextRetryAt: now, PayloadVersion: tradeEventPayloadVersion, Payload: "{}", CreateTimes: now, UpdateTimes: now})
		return err
	}); err != nil {
		return err
	}
	if isSeconds && l.svcCtx.DelayQueue != nil {
		if err := enqueueSecondsActivation(l.svcCtx, order); err != nil {
			// The durable seconds row is already ACTIVATING. The minute-level
			// recovery job will pick it up if the delay queue is unavailable.
			l.Errorf("enqueue seconds activation failed, orderId=%d err=%v", order.Id, err)
		}
	}
	if err := enqueueOrderExpiration(l.svcCtx, order); err != nil {
		l.Errorf("enqueue order expiration failed, orderId=%d err=%v", order.Id, err)
	}
	return nil
}

func (l *PlaceOrderLogic) markAssetReservationRetry(order *models.TTradeOrder, cause error) error {
	reservation, err := l.svcCtx.TradeAssetReservationModel.FindOneByTenantIdReservationNo(l.ctx, order.TenantId, order.OrderNo)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	reservation.RetryCount++
	reservation.NextRetryAt = utils.NowMillis() + 1000
	reservation.LastErrorMsg = cause.Error()
	reservation.UpdateTimes = utils.NowMillis()
	return l.svcCtx.TradeAssetReservationModel.Update(l.ctx, reservation)
}

func (l *PlaceOrderLogic) rejectOrderAfterFreezeFailure(order *models.TTradeOrder, plan *placeOrderPlan, cause error) error {
	now := utils.NowMillis()
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		reservationModel := models.NewTTradeAssetReservationModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis)

		order.Status = int64(trade.OrderStatus_ORDER_STATUS_REJECTED)
		order.CancelReason = fmt.Sprintf("asset freeze rejected: %v", cause)
		order.UpdateTimes = now
		if err := orderModel.Update(ctx, order); err != nil {
			return err
		}
		if plan.frozenAmount.IsPositive() {
			reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, order.TenantId, order.OrderNo)
			if err != nil {
				return err
			}
			reservation.Status = int64(trade.AssetReservationStatus_ASSET_RESERVATION_STATUS_FAILED)
			reservation.LastErrorMsg = cause.Error()
			reservation.NextRetryAt = 0
			reservation.UpdateTimes = now
			if err := reservationModel.Update(ctx, reservation); err != nil {
				return err
			}
		}
		if plan.reservedCloseQty.IsPositive() {
			if err := positionModel.ReleaseCloseQty(ctx, plan.positionID, plan.reservedCloseQty, now); err != nil {
				return err
			}
		}
		_, err := eventModel.Insert(ctx, &models.TBizTradeEvent{TenantId: order.TenantId, EventNo: order.OrderNo + "-REJECTED", EventType: "ORDER_REJECTED", BizId: order.OrderNo, BizType: "order", UserId: order.UserId, SymbolId: order.SymbolId, ProductType: order.ProductType, OperatorId: order.UserId, Source: int64(trade.SourceType_SOURCE_TYPE_USER), EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20, NextRetryAt: now, Payload: "{}", CreateTimes: now, UpdateTimes: now})
		return err
	})
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

// placeOrderPlan is the immutable result of validation and reservation
// calculation. Database writes must use this result instead of recalculating
// product rules in the transaction.
type placeOrderPlan struct {
	price            decimal.Decimal
	qty              decimal.Decimal
	notional         decimal.Decimal
	riskPrice        decimal.Decimal
	frozenAsset      string
	frozenAmount     decimal.Decimal
	marginAmount     decimal.Decimal
	reservedCloseQty decimal.Decimal
	riskTierID       int64
	positionID       int64
	positionVersion  int64
}

func (l *PlaceOrderLogic) preparePlaceOrder(
	tenantID, userID int64,
	symbol *models.TTradeSymbol,
	secondsCfg *models.TTradeSymbolSeconds,
	in *trade.PlaceOrderReq,
	orderType trade.OrderType,
	triggerKind trade.TriggerKind,
	timeInForce trade.TimeInForce,
	leverage int64,
	price, qty, amount decimal.Decimal,
) (*placeOrderPlan, error) {
	if symbol.Status == int64(trade.SymbolStatus_SYMBOL_STATUS_DISABLED) || symbol.Status == int64(trade.SymbolStatus_SYMBOL_STATUS_UNKNOWN) {
		return nil, errors.New("symbol is not tradable")
	}
	now := nowMillis()
	if (symbol.ListingTime > 0 && now < symbol.ListingTime) || (symbol.TradingStartTime > 0 && now < symbol.TradingStartTime) || (symbol.TradingEndTime > 0 && now >= symbol.TradingEndTime) {
		return nil, errors.New("symbol is outside its trading time")
	}
	if symbol.Status == int64(trade.SymbolStatus_SYMBOL_STATUS_CLOSE_ONLY) && !isDerivativeProduct(common.ProductType(symbol.ProductType)) {
		return nil, errors.New("close-only is only valid for derivative symbols")
	}
	if symbol.Status == int64(trade.SymbolStatus_SYMBOL_STATUS_CLOSE_ONLY) && in.IsReduceOnly != common.YesNo_YES_NO_YES {
		return nil, errors.New("symbol only accepts reduce-only orders")
	}
	if err := l.validateUserTradingEnabled(tenantID, userID, symbol); err != nil {
		return nil, err
	}

	plan := &placeOrderPlan{price: price, qty: qty, notional: amount, riskPrice: price}
	if symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_SECONDS) {
		if err := validateSymbolNotional(symbol, amount); err != nil {
			return nil, err
		}
		risk, err := NewCheckOrderRiskLogic(l.ctx, l.svcCtx).CheckOrderRisk(&trade.CheckOrderRiskReq{TenantId: tenantID, UserId: userID, SymbolId: symbol.Id, Amount: amount.String()})
		if err != nil {
			return nil, err
		}
		if risk.Passed != 1 {
			return nil, fmt.Errorf("order rejected by risk control: %s", risk.RejectMsg)
		}
		plan.frozenAsset, plan.frozenAmount = symbol.SettleAsset, amount
		return plan, nil
	}
	if in.Side != common.Side_SIDE_BUY && in.Side != common.Side_SIDE_SELL {
		return nil, errors.New("invalid order side")
	}
	// Reject derivative-only fields before quantity validation or reference-price
	// lookup. Otherwise a malformed spot request can be reported as a market-data
	// failure, which hides the actual request error.
	if symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_SPOT) &&
		(in.IsReduceOnly == common.YesNo_YES_NO_YES ||
			in.PositionSide != trade.PositionSide_POSITION_SIDE_UNKNOWN ||
			in.MarginMode != trade.MarginMode_MARGIN_MODE_UNKNOWN ||
			in.Leverage > 0) {
		return nil, errors.New("position, margin and reduce-only fields are not valid for spot orders")
	}
	if err := validateSymbolOrderIncrements(symbol, orderType, price, qty); err != nil {
		return nil, err
	}
	if orderType == trade.OrderType_ORDER_TYPE_MARKET {
		if timeInForce != trade.TimeInForce_TIME_IN_FORCE_IOC && timeInForce != trade.TimeInForce_TIME_IN_FORCE_FOK {
			return nil, errors.New("market order only supports IOC or FOK")
		}
		if symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_SPOT) && in.Side == common.Side_SIDE_BUY {
			if !amount.IsPositive() {
				return nil, errors.New("spot market buy requires quote amount")
			}
		} else if !qty.IsPositive() {
			return nil, errors.New("market sell and derivative market order require quantity")
		}
	}
	if !plan.riskPrice.IsPositive() && qty.IsPositive() {
		var err error
		plan.riskPrice, err = l.bestOppositePrice(tenantID, symbol, in.Side)
		if err != nil {
			return nil, err
		}
	}

	switch common.ProductType(symbol.ProductType) {
	case common.ProductType_PRODUCT_TYPE_SPOT:
		cfg, err := l.svcCtx.TradeSymbolSpotModel.FindOneByTenantIdSymbolId(l.ctx, symbol.TenantId, symbol.Id)
		if errors.Is(err, models.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("spot symbol configuration not found")
		}
		if err != nil {
			return nil, err
		}
		if (in.Side == common.Side_SIDE_BUY && cfg.BuyEnabled != int64(common.Enable_ENABLE_ENABLED)) || (in.Side == common.Side_SIDE_SELL && cfg.SellEnabled != int64(common.Enable_ENABLE_ENABLED)) {
			return nil, errors.New("spot side is disabled")
		}
		if !plan.notional.IsPositive() && qty.IsPositive() {
			plan.notional = tradeMinorAmountAtPrice(plan.riskPrice, qty)
		}
		if err := validateSymbolNotional(symbol, plan.notional); err != nil {
			return nil, err
		}
		plan.frozenAsset, plan.frozenAmount = spotFrozenAssetAndAmount(symbol, in.Side, qty, plan.notional)
		if in.Side == common.Side_SIDE_BUY {
			plan.frozenAmount = plan.frozenAmount.Add(plan.notional.Mul(cfg.TakerFeeRate))
		}

	case common.ProductType_PRODUCT_TYPE_DERIVATIVE:
		cfg, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, symbol.TenantId, symbol.Id)
		if err != nil {
			return nil, err
		}
		// Cross margin is account-level risk. It cannot be accepted until wallet
		// equity, open-order margin and every position sharing the margin asset are
		// continuously projected from Asset and mark-price events. Silently using
		// the isolated-position formula here would permit orders that cannot be
		// liquidated correctly.
		if in.MarginMode == trade.MarginMode_MARGIN_MODE_CROSS {
			return nil, errors.New("cross margin is temporarily unavailable: account-level risk projection is not enabled")
		}
		if in.MarginMode != trade.MarginMode_MARGIN_MODE_ISOLATED {
			return nil, errors.New("invalid derivative margin mode")
		}
		if symbol.ContractType != int64(common.ContractType_CONTRACT_TYPE_PERPETUAL) && symbol.ContractType != int64(common.ContractType_CONTRACT_TYPE_DELIVERY) {
			return nil, errors.New("invalid derivative contract type")
		}
		if symbol.ContractValueType != int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR) && symbol.ContractValueType != int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE) {
			return nil, errors.New("invalid derivative contract value type")
		}
		if symbol.ContractType == int64(common.ContractType_CONTRACT_TYPE_DELIVERY) {
			if cfg.DeliveryTime <= now || cfg.MatchingStopTime <= now {
				return nil, errors.New("delivery contract no longer accepts orders")
			}
			if in.IsReduceOnly != common.YesNo_YES_NO_YES && cfg.OpenCutoffTime > 0 && now >= cfg.OpenCutoffTime {
				return nil, errors.New("delivery contract has stopped opening positions")
			}
		}
		if !qty.IsPositive() || !plan.riskPrice.IsPositive() || leverage <= 0 {
			return nil, errors.New("derivative order requires quantity, risk price and leverage")
		}
		if cfg.SupportIsolated != 1 {
			return nil, errors.New("margin mode is not supported")
		}
		if in.PositionSide != trade.PositionSide_POSITION_SIDE_NET && in.PositionSide != trade.PositionSide_POSITION_SIDE_LONG && in.PositionSide != trade.PositionSide_POSITION_SIDE_SHORT {
			return nil, errors.New("invalid derivative position side")
		}
		if err := validateContractSideSwitch(cfg, in); err != nil {
			return nil, err
		}
		values, err := contractmath.CalculateTradeValues(symbol.ContractValueType, qty, cfg.ContractSize, plan.riskPrice)
		if err != nil {
			return nil, err
		}
		plan.notional = values.QuoteNotional
		plan.marginAmount, err = contractmath.CalculateMargin(values, leverage)
		if err != nil {
			return nil, err
		}
		if err := validateSymbolNotional(symbol, plan.notional); err != nil {
			return nil, err
		}
		tier, err := l.svcCtx.ContractRiskLimitTierModel.FindByNotional(l.ctx, tenantID, symbol.Id, plan.notional)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if tier != nil {
			if leverage > tier.MaxLeverage {
				return nil, errors.New("leverage exceeds risk tier maximum")
			}
			plan.riskTierID = tier.Id
		}
		fee := contractmath.CalculateFee(values, cfg.TakerFeeRate)
		plan.frozenAsset = marginAssetForSymbol(symbol)
		isHedgeClose := in.PositionSide != trade.PositionSide_POSITION_SIDE_NET && isClosingFill(int64(in.PositionSide), int64(in.Side))
		if in.IsReduceOnly == common.YesNo_YES_NO_YES || isHedgeClose {
			lookupSide := int64(in.PositionSide)
			if in.PositionSide == trade.PositionSide_POSITION_SIDE_NET {
				lookupSide, _ = netPositionSides(int64(in.Side))
			}
			position, findErr := l.svcCtx.ContractPositionModel.FindOneByTenantIdUserIdSymbolIdPositionSideMarginMode(l.ctx, tenantID, userID, symbol.Id, lookupSide, int64(in.MarginMode))
			if findErr != nil {
				return nil, errors.New("no matching position available to close")
			}
			if position.Status != 1 || position.AvailQty.LessThan(qty) {
				return nil, errors.New("insufficient available position quantity")
			}
			plan.marginAmount = decimal.Zero
			plan.reservedCloseQty = qty
			plan.positionID = position.Id
			plan.positionVersion = position.Version
			plan.frozenAmount = fee
		} else {
			plan.frozenAmount = plan.marginAmount.Add(fee)
		}
	default:
		return nil, errors.New("unsupported product type")
	}

	risk, err := NewCheckOrderRiskLogic(l.ctx, l.svcCtx).CheckOrderRisk(&trade.CheckOrderRiskReq{TenantId: tenantID, UserId: userID, SymbolId: symbol.Id, Side: in.Side, Price: plan.riskPrice.String(), Qty: qty.String(), Amount: plan.notional.String()})
	if err != nil {
		return nil, err
	}
	if risk.Passed != 1 {
		return nil, fmt.Errorf("order rejected by risk control: %s", risk.RejectMsg)
	}
	_ = triggerKind
	_ = secondsCfg
	return plan, nil
}

func nowMillis() int64 { return timeNow().UnixMilli() }

var timeNow = func() time.Time { return time.Now() }

func (l *PlaceOrderLogic) validateUserTradingEnabled(tenantID, userID int64, symbol *models.TTradeSymbol) error {
	for _, symbolID := range []int64{symbol.Id, 0} {
		cfg, err := l.svcCtx.TradeUserConfigModel.FindOneByTenantIdUserIdProductTypeSymbolId(l.ctx, tenantID, userID, symbol.ProductType, symbolID)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if cfg != nil && cfg.TradeEnabled != int64(common.Enable_ENABLE_ENABLED) {
			return errors.New("user trading is disabled")
		}
	}
	return nil
}

func validateSymbolOrderIncrements(symbol *models.TTradeSymbol, orderType trade.OrderType, price, qty decimal.Decimal) error {
	if orderType == trade.OrderType_ORDER_TYPE_LIMIT {
		if decimalPlaces(price) > symbol.PriceScale || symbol.MinPrice.IsPositive() && price.LessThan(symbol.MinPrice) || symbol.MaxPrice.IsPositive() && price.GreaterThan(symbol.MaxPrice) || symbol.PriceTick.IsPositive() && !price.Mod(symbol.PriceTick).IsZero() {
			return errors.New("price violates symbol price limit or tick size")
		}
	}
	if qty.IsPositive() && (decimalPlaces(qty) > symbol.QtyScale || symbol.MinQty.IsPositive() && qty.LessThan(symbol.MinQty) || symbol.MaxQty.IsPositive() && qty.GreaterThan(symbol.MaxQty) || symbol.QtyStep.IsPositive() && !qty.Mod(symbol.QtyStep).IsZero()) {
		return errors.New("quantity violates symbol quantity limit or step size")
	}
	return nil
}

func decimalPlaces(value decimal.Decimal) int64 {
	text := value.String()
	for index, char := range text {
		if char == '.' {
			return int64(len(text) - index - 1)
		}
	}
	return 0
}

func validateSymbolNotional(symbol *models.TTradeSymbol, notional decimal.Decimal) error {
	if !notional.IsPositive() || symbol.MinNotional.IsPositive() && notional.LessThan(symbol.MinNotional) || symbol.MaxNotional.IsPositive() && notional.GreaterThan(symbol.MaxNotional) {
		return errors.New("notional violates symbol limits")
	}
	return nil
}

func validateContractSideSwitch(cfg *models.TTradeSymbolContract, in *trade.PlaceOrderReq) error {
	reduce := in.IsReduceOnly == common.YesNo_YES_NO_YES
	if reduce {
		if in.PositionSide == trade.PositionSide_POSITION_SIDE_LONG && (in.Side != common.Side_SIDE_SELL || cfg.CloseLongEnabled != 1) || in.PositionSide == trade.PositionSide_POSITION_SIDE_SHORT && (in.Side != common.Side_SIDE_BUY || cfg.CloseShortEnabled != 1) {
			return errors.New("close direction is invalid or disabled")
		}
		return nil
	}
	if in.Side == common.Side_SIDE_BUY && cfg.OpenLongEnabled != 1 || in.Side == common.Side_SIDE_SELL && cfg.OpenShortEnabled != 1 {
		return errors.New("open direction is disabled")
	}
	return nil
}

func (l *PlaceOrderLogic) bestOppositePrice(tenantID int64, symbol *models.TTradeSymbol, side common.Side) (decimal.Decimal, error) {
	opposite := int64(common.Side_SIDE_SELL)
	if side == common.Side_SIDE_SELL {
		opposite = int64(common.Side_SIDE_BUY)
	}
	orders, err := l.svcCtx.TradeOrderModel.FindOpenMatchOrders(l.ctx, tenantID, symbol.Id, symbol.ProductType, opposite, matchableOrderStatuses(), int64(trade.OrderType_ORDER_TYPE_MARKET), 1)
	if err != nil {
		return decimal.Zero, err
	}
	if len(orders) > 0 && orders[0].Price.IsPositive() {
		return orders[0].Price, nil
	}

	const maxReferencePriceAge = int64(30_000)
	if isDerivativeProduct(common.ProductType(symbol.ProductType)) {
		contract, contractErr := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, symbol.TenantId, symbol.Id)
		if contractErr != nil && !errors.Is(contractErr, models.ErrNotFound) {
			return decimal.Zero, contractErr
		}
		if contract != nil && contract.MarkPriceSource != "" {
			quote, quoteErr := NewProcessSecondsSettlementsLogic(l.ctx, l.svcCtx).getValidQuoteKind("MARK_PRICE", contract.MarkPriceSource, symbol.Id, maxReferencePriceAge)
			if quoteErr == nil {
				price := mustParseFloat(quote.LastPrice)
				if price.IsPositive() {
					return price, nil
				}
			}
			// A newly listed contract may receive authoritative source quotes before
			// its versioned price formula has emitted the first MARK snapshot.
			// Use the same configured source's confirmed FINAL_QUOTE during that
			// bootstrap window; the 30-second validity bound still applies.
			quote, quoteErr = NewProcessSecondsSettlementsLogic(l.ctx, l.svcCtx).getValidQuoteKind("FINAL_QUOTE", contract.MarkPriceSource, symbol.Id, maxReferencePriceAge)
			if quoteErr == nil {
				price := mustParseFloat(quote.LastPrice)
				if price.IsPositive() {
					return price, nil
				}
			}
		}
	}
	if symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_SPOT) {
		if source := l.spotReferencePriceSource(symbol); source != "" {
			quote, quoteErr := NewProcessSecondsSettlementsLogic(l.ctx, l.svcCtx).getValidQuoteKind("FINAL_QUOTE", source, symbol.Id, maxReferencePriceAge)
			if quoteErr == nil {
				price := mustParseFloat(quote.LastPrice)
				if price.IsPositive() {
					return price, nil
				}
			}
		}
	}

	snapshot, snapshotErr := l.svcCtx.TradeMarketSnapshotModel.FindLatestConfirmed(l.ctx, tenantID, symbol.Id, nowMillis()-maxReferencePriceAge)
	if snapshotErr != nil {
		if errors.Is(snapshotErr, models.ErrNotFound) {
			return decimal.Zero, errors.New("no valid reference price for market order")
		}
		return decimal.Zero, snapshotErr
	}
	if isDerivativeProduct(common.ProductType(symbol.ProductType)) && snapshot.MarkPrice.IsPositive() {
		return snapshot.MarkPrice, nil
	}
	if snapshot.Price.IsPositive() {
		return snapshot.Price, nil
	}
	if snapshot.IndexPrice.IsPositive() {
		return snapshot.IndexPrice, nil
	}
	return decimal.Zero, errors.New("no valid reference price for market order")
}

func (l *PlaceOrderLogic) spotReferencePriceSource(symbol *models.TTradeSymbol) string {
	secondsSymbol, err := l.svcCtx.TradeSymbolModel.FindOneByTenantIdSymbolProductTypeContractTypeContractValueType(
		l.ctx, symbol.TenantId, symbol.Symbol, int64(common.ProductType_PRODUCT_TYPE_SECONDS), 0, 0,
	)
	if err == nil {
		configs, configErr := l.svcCtx.TradeSymbolSecondsModel.FindAllByTenantIdSymbolId(l.ctx, symbol.TenantId, secondsSymbol.Id)
		if configErr == nil {
			for _, config := range configs {
				if source := strings.TrimSpace(config.StartPriceSource); source != "" {
					return source
				}
				if source := strings.TrimSpace(config.SettlementPriceSource); source != "" {
					return source
				}
			}
		}
	}
	for _, variant := range [][2]int64{{1, 1}, {1, 2}, {2, 1}, {2, 2}} {
		contractSymbol, findErr := l.svcCtx.TradeSymbolModel.FindOneByTenantIdSymbolProductTypeContractTypeContractValueType(
			l.ctx, symbol.TenantId, symbol.Symbol, int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE), variant[0], variant[1],
		)
		if findErr != nil {
			continue
		}
		config, configErr := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, symbol.TenantId, contractSymbol.Id)
		if configErr == nil {
			if source := strings.TrimSpace(config.MarkPriceSource); source != "" {
				return source
			}
		}
	}
	return ""
}
