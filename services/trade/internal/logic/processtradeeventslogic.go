package logic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/common/conv"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

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
		if err := l.retryTradeEvents(in); err != nil {
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

type triggerPriceKey struct {
	tenantId   int64
	symbolId   int64
	marketType int64
}

func (l *ProcessTradeEventsLogic) retryTradeEvents(in *trade.TradeTaskReq) error {
	now := utils.NowMillis()
	cursor := int64(0)
	for {
		items, _, err := l.svcCtx.BizTradeEventModel.FindPage(l.ctx, models.BizTradeEventPageFilter{
			TenantId:    in.GetTenantId(),
			EventStatus: int64(trade.EventStatus_EVENT_STATUS_FAILED),
			TimeEnd:     now,
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for _, item := range items {
			cursor = item.Id
			if item.MaxRetryCount > 0 && item.RetryCount >= item.MaxRetryCount {
				continue
			}
			if item.NextRetryAt > 0 && item.NextRetryAt > now {
				continue
			}
			item.EventStatus = int64(trade.EventStatus_EVENT_STATUS_PENDING)
			item.RetryCount++
			item.NextRetryAt = now
			item.LastErrorMsg = ""
			item.UpdateTimes = now
			if err := l.svcCtx.BizTradeEventModel.Update(l.ctx, item); err != nil {
				return err
			}
		}
		if len(items) < 100 {
			return nil
		}
	}
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
			if _, err := l.rejectFreezingOrderIfNeeded(order.Id, now); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessTradeEventsLogic) rejectFreezingOrderIfNeeded(orderID, now int64) (*models.TTradeOrder, error) {
	var rejectedOrder *models.TTradeOrder
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		order, err := orderModel.FindOneForUpdate(ctx, orderID)
		if err != nil {
			return err
		}
		if !shouldRecoverFreezingOrder(order, now) {
			return nil
		}
		order.Status = int64(trade.OrderStatus_ORDER_STATUS_REJECTED)
		order.CancelReason = "rejected by freeze timeout"
		order.UpdateTimes = now
		if err := orderModel.Update(ctx, order); err != nil {
			return err
		}
		rejectedOrder = order
		return nil
	})
	if err != nil || rejectedOrder == nil {
		return rejectedOrder, err
	}
	return rejectedOrder, removeOrderBookOrder(l.svcCtx, l.ctx, rejectedOrder)
}

func (l *ProcessTradeEventsLogic) triggerWaitingOrders(in *trade.TradeTaskReq) error {
	now := utils.NowMillis()
	cursor := int64(0)
	priceCache := make(map[triggerPriceKey]float64)
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
			key := triggerPriceKey{tenantId: order.TenantId, symbolId: order.SymbolId, marketType: order.MarketType}
			triggerPrice, ok := priceCache[key]
			if !ok {
				triggerPrice, err = l.svcCtx.TradeFillModel.FindLastPrice(l.ctx, order.TenantId, order.SymbolId, order.MarketType)
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

func (l *ProcessTradeEventsLogic) triggerOrderIfNeeded(orderID int64, triggerPrice float64, now int64) error {
	var triggeredOrder *models.TTradeOrder
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
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
		triggeredOrder = order
		return nil
	})
	if err != nil || triggeredOrder == nil {
		return err
	}
	return cacheOrderBookOrder(l.svcCtx, l.ctx, triggeredOrder)
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
				if err := removeOrderBookOrder(l.svcCtx, l.ctx, expiredOrder); err != nil {
					return err
				}
				if err := unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, expiredOrder, "trade expired order unfreeze"); err != nil {
					return err
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
		order.Status = int64(trade.OrderStatus_ORDER_STATUS_EXPIRED)
		order.CancelReason = orderExpireReason(order)
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
	return expiredOrder, removeOrderBookOrder(l.svcCtx, l.ctx, expiredOrder)
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
