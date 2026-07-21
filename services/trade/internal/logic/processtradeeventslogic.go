package logic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/common/conv"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/realtime"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessTradeEventsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessTradeEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessTradeEventsLogic {
	return &ProcessTradeEventsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 交易事件处理（失败重试/订单过期/冻结资产修复）
func (l *ProcessTradeEventsLogic) ProcessTradeEvents(in *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
	return runTradeTaskWithLock(l.ctx, l.svcCtx, "process_trade_events", func() (*trade.TradeTaskResp, error) {
		if err := NewProcessFillSettlementsLogic(l.ctx, l.svcCtx).Process(in.GetTenantId()); err != nil {
			return nil, err
		}
		if err := NewProcessReservationReleasesLogic(l.ctx, l.svcCtx).Process(in.GetTenantId()); err != nil {
			return nil, err
		}
		if err := l.recoverTerminatingOrders(in); err != nil {
			return nil, err
		}
		if err := l.recoverSettlementPendingOrders(in); err != nil {
			return nil, err
		}
		if err := l.dispatchPendingTradeEvents(in); err != nil {
			return nil, err
		}
		if err := l.recoverFreezingOrders(in); err != nil {
			return nil, err
		}
		if err := l.triggerWaitingOrders(in); err != nil {
			return nil, err
		}
		if err := l.expireImmediateOrders(in); err != nil {
			return nil, err
		}
		if err := l.repairFrozenAssets(in); err != nil {
			return nil, err
		}
		return okTradeTaskResp(), nil
	})
}

func (l *ProcessTradeEventsLogic) recoverSettlementPendingOrders(in *trade.TradeTaskReq) error {
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{TenantId: in.GetTenantId(), Statuses: []int64{int64(trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING)}}, cursor, 100)
		if err != nil {
			return err
		}
		for _, order := range orders {
			cursor = order.Id
			if err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				return finalizeSettledOrder(ctx, sqlx.NewSqlConnFromSession(session), l.svcCtx, order.Id, utils.NowMillis())
			}); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessTradeEventsLogic) recoverTerminatingOrders(in *trade.TradeTaskReq) error {
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{TenantId: in.GetTenantId(), Statuses: terminatingOrderStatuses()}, cursor, 100)
		if err != nil {
			return err
		}
		for _, order := range orders {
			cursor = order.Id
			if err := unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, order, "trade termination recovery release"); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

// dispatchPendingTradeEvents republishes all durable outbox records. Publishing
// does not mark the record successful; only the consumer may acknowledge it.
// This makes Redis Pub/Sub the low-latency path while the outbox remains the
// recovery source when a process or message is lost.
func (l *ProcessTradeEventsLogic) dispatchPendingTradeEvents(in *trade.TradeTaskReq) error {
	cursor := int64(0)
	for {
		now := utils.NowMillis()
		items, err := l.svcCtx.BizTradeEventModel.FindDispatchable(l.ctx, in.GetTenantId(), now, now-realtime.ClaimLeaseMillis, cursor, 100, nil)
		if err != nil {
			return err
		}
		for _, item := range items {
			cursor = item.Id
			event := realtime.Event{Version: item.PayloadVersion, Consumer: item.Consumer, EventNo: item.EventNo, Type: item.EventType, TenantID: item.TenantId, BizID: item.BizId, Payload: item.Payload}
			if err := publishTradeOutboxEvent(l.ctx, l.svcCtx, event); err != nil {
				l.Errorf("dispatch trade event failed, eventNo=%s: %v", item.EventNo, err)
			}
		}
		if len(items) < 100 {
			return nil
		}
	}
}

type triggerPriceKey struct {
	tenantId    int64
	symbolId    int64
	productType int64
}

func (l *ProcessTradeEventsLogic) recoverFreezingOrders(in *trade.TradeTaskReq) error {
	now := utils.NowMillis()
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
			TenantId: in.GetTenantId(),
			Statuses: freezingOrderStatuses(),
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, order := range orders {
			cursor = order.Id
			if err := l.recoverFreezingOrder(order, now); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessTradeEventsLogic) recoverFreezingOrder(order *models.TTradeOrder, now int64) error {
	if !shouldRecoverFreezingOrder(order, now) {
		return nil
	}
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, order.SymbolId)
	if err != nil {
		return err
	}
	reservation, err := l.svcCtx.TradeAssetReservationModel.FindOneByTenantIdReservationNo(l.ctx, order.TenantId, order.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	var assetName string
	var amount decimal.Decimal
	if reservation != nil {
		if reservation.NextRetryAt > now {
			return nil
		}
		assetName, amount = reservation.Asset, reservation.ReservedAmount
	}
	freezeNo, freezeErr := freezeOrderAsset(l.svcCtx, l.ctx, order, symbol, assetName, amount)
	placeLogic := NewPlaceOrderLogic(l.ctx, l.svcCtx)
	if freezeErr != nil {
		if !isDefinitiveAssetFreezeError(freezeErr) {
			return placeLogic.markAssetReservationRetry(order, freezeErr)
		}
		plan, err := l.recoveryRejectPlan(order)
		if err != nil {
			return err
		}
		plan.frozenAmount = amount
		return placeLogic.rejectOrderAfterFreezeFailure(order, plan, freezeErr)
	}
	return placeLogic.finalizeAcceptedOrder(order, freezeNo, amount, trade.TriggerKind(order.TriggerKind), trade.OrderType(order.OrderType), order.TriggerPrice, order.ProductType == int64(trade.ProductType_PRODUCT_TYPE_SECONDS))
}

func (l *ProcessTradeEventsLogic) recoveryRejectPlan(order *models.TTradeOrder) (*placeOrderPlan, error) {
	plan := &placeOrderPlan{}
	if order.ProductType != int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE) || order.IsReduceOnly != int64(common.YesNo_YES_NO_YES) {
		return plan, nil
	}
	ext, err := l.svcCtx.TradeOrderContractModel.FindOneByTenantIdOrderId(l.ctx, order.TenantId, order.Id)
	if err != nil {
		return nil, err
	}
	position, err := l.svcCtx.ContractPositionModel.FindOneByTenantIdUserIdSymbolIdPositionSideMarginMode(l.ctx, order.TenantId, order.UserId, order.SymbolId, order.PositionSide, ext.MarginMode)
	if err != nil {
		return nil, err
	}
	plan.positionID = position.Id
	plan.reservedCloseQty = ext.ReservedCloseQty
	return plan, nil
}

func (l *ProcessTradeEventsLogic) triggerWaitingOrders(in *trade.TradeTaskReq) error {
	now := utils.NowMillis()
	cursor := int64(0)
	priceCache := make(map[triggerPriceKey]decimal.Decimal)
	for {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
			TenantId: in.GetTenantId(),
			Statuses: triggerWaitingOrderStatuses(),
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, order := range orders {
			cursor = order.Id
			key := triggerPriceKey{tenantId: order.TenantId, symbolId: order.SymbolId, productType: order.ProductType}
			triggerPrice, ok := priceCache[key]
			if !ok {
				triggerPrice, err = l.svcCtx.TradeFillModel.FindLastPrice(l.ctx, order.TenantId, order.SymbolId, order.ProductType)
				if errors.Is(err, models.ErrNotFound) {
					continue
				}
				if err != nil {
					return err
				}
				priceCache[key] = triggerPrice
			}
			if !shouldTriggerOrder(order, triggerPrice) {
				continue
			}
			if err := l.triggerOrderIfNeeded(order.Id, triggerPrice, now); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessTradeEventsLogic) triggerOrderIfNeeded(orderID int64, triggerPrice decimal.Decimal, now int64) error {
	var triggeredOrder *models.TTradeOrder
	var eventNo string
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis)
		order, err := orderModel.FindOneForUpdate(ctx, orderID)
		if err != nil {
			return err
		}
		if !shouldTriggerOrder(order, triggerPrice) {
			return nil
		}
		ext, err := parseOrderAssetExt(conv.NullStringValue(order.BizExt))
		if err != nil {
			return err
		}
		if ext.OriginalOrderType == 0 {
			ext.OriginalOrderType = order.OrderType
		}
		ext.TriggeredAt = now
		ext.TriggerPrice = conv.FloatString(triggerPrice)
		ext.TriggerSource = "last_price"
		extValue, err := marshalOrderAssetExt(ext)
		if err != nil {
			return err
		}
		order.BizExt = sql.NullString{String: extValue, Valid: extValue != ""}
		order.OrderType = triggeredOrderExecutionType(order)
		order.TimeInForce = triggeredTimeInForce(order)
		order.Status = int64(trade.OrderStatus_ORDER_STATUS_PENDING)
		order.UpdateTimes = now
		if err := orderModel.Update(ctx, order); err != nil {
			return err
		}
		eventNo = derivedTradeBizNo(order.OrderNo, "TRIGGERED")
		if _, err := eventModel.Insert(ctx, &models.TBizTradeEvent{
			TenantId: order.TenantId, EventNo: eventNo, EventType: realtime.EventOrderAccepted,
			BizId: order.OrderNo, BizType: "order", UserId: order.UserId, SymbolId: order.SymbolId,
			ProductType: order.ProductType, OperatorId: order.UserId, Source: int64(trade.SourceType_SOURCE_TYPE_SYSTEM),
			Consumer: tradeEventConsumer(realtime.EventOrderAccepted), EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING), MaxRetryCount: 20, NextRetryAt: now,
			PayloadVersion: tradeEventPayloadVersion,
			Payload:        "{}", CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		triggeredOrder = order
		return nil
	})
	if err != nil || triggeredOrder == nil {
		return err
	}
	if err := cacheOrderBookOrder(l.svcCtx, l.ctx, triggeredOrder); err != nil {
		return err
	}
	event := realtime.Event{EventNo: eventNo, Type: realtime.EventOrderAccepted, TenantID: triggeredOrder.TenantId, BizID: triggeredOrder.OrderNo, OrderID: triggeredOrder.Id}
	if err := publishTradeOutboxEvent(l.ctx, l.svcCtx, event); err != nil {
		l.Errorf("publish triggered order event failed, orderId=%d eventNo=%s err=%v", triggeredOrder.Id, eventNo, err)
	}
	return nil
}

func (l *ProcessTradeEventsLogic) expireImmediateOrders(in *trade.TradeTaskReq) error {
	now := utils.NowMillis()
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
			TenantId: in.GetTenantId(),
			Statuses: matchableOrderStatuses(),
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, order := range orders {
			cursor = order.Id
			expiredOrder, err := l.expireOrderIfNeeded(order.Id, now)
			if err != nil {
				return err
			}
			if expiredOrder != nil {
				if err := unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, expiredOrder, "trade expired order unfreeze"); err != nil {
					return err
				}
				if err := removeOrderBookOrder(l.svcCtx, l.ctx, expiredOrder); err != nil {
					l.Errorf("remove expired order from cache failed, orderId=%d err=%v", expiredOrder.Id, err)
				}
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessTradeEventsLogic) expireOrderIfNeeded(orderID, now int64) (*models.TTradeOrder, error) {
	var expiredOrder *models.TTradeOrder
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		order, err := orderModel.FindOneForUpdate(ctx, orderID)
		if err != nil {
			return err
		}
		if !shouldExpireOrder(order, now) {
			return nil
		}
		order.Status = int64(trade.OrderStatus_ORDER_STATUS_EXPIRING)
		order.CanceledQty = decimalMaxZero(order.Qty.Sub(order.FilledQty))
		order.CancelReason = orderExpireReason(order)
		order.Version++
		order.UpdateTimes = now
		if err := orderModel.Update(ctx, order); err != nil {
			return err
		}
		expiredOrder = order
		return nil
	})
	if err != nil || expiredOrder == nil {
		return expiredOrder, err
	}
	return expiredOrder, nil
}

func (l *ProcessTradeEventsLogic) repairFrozenAssets(in *trade.TradeTaskReq) error {
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
			TenantId: in.GetTenantId(),
			Statuses: []int64{
				int64(trade.OrderStatus_ORDER_STATUS_CANCELED),
				int64(trade.OrderStatus_ORDER_STATUS_REJECTED),
				int64(trade.OrderStatus_ORDER_STATUS_EXPIRED),
			},
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, order := range orders {
			cursor = order.Id
			if err := unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, order, "trade frozen asset repair unfreeze"); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}
