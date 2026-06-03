package logic

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"
)

const (
	orderBookKeyPrefix   = "trade:book"
	orderBookMarketScore = -1e15
	orderBookScanFactor  = int64(3)
	orderBookRestoreSize = int64(500)
)

func orderBookKey(order *models.TTradeOrder) string {
	if order == nil {
		return ""
	}
	return orderBookKeyBySide(order.TenantId, order.SymbolId, order.MarketType, order.Side)
}

func orderBookKeyBySide(tenantID, symbolID, marketType, side int64) string {
	sideName := "sell"
	if side == int64(trade.TradeSide_TRADE_SIDE_BUY) {
		sideName = "buy"
	}
	return fmt.Sprintf("%s:%d:%d:%d:%s", orderBookKeyPrefix, tenantID, marketType, symbolID, sideName)
}

func orderBookMember(orderID int64) string {
	return fmt.Sprintf("%020d", orderID)
}

func orderBookMemberID(member string) (int64, error) {
	return strconv.ParseInt(member, 10, 64)
}

func orderBookScore(order *models.TTradeOrder) float64 {
	if order == nil {
		return 0
	}
	if order.OrderType == int64(trade.OrderType_ORDER_TYPE_MARKET) {
		return orderBookMarketScore
	}
	if order.Side == int64(trade.TradeSide_TRADE_SIDE_BUY) {
		return -order.Price
	}
	return order.Price
}

func isOrderBookOrder(order *models.TTradeOrder) bool {
	return order != nil &&
		isMatchableOrderStatus(order.Status) &&
		(order.OrderType == int64(trade.OrderType_ORDER_TYPE_MARKET) ||
			order.OrderType == int64(trade.OrderType_ORDER_TYPE_LIMIT))
}

func cacheOrderBookOrder(svcCtx *svc.ServiceContext, ctx context.Context, order *models.TTradeOrder) error {
	if svcCtx == nil || svcCtx.Redis == nil || order == nil || !isOrderBookOrder(order) {
		return nil
	}
	_, err := svcCtx.Redis.ZaddFloatCtx(ctx, orderBookKey(order), orderBookScore(order), orderBookMember(order.Id))
	return err
}

func removeOrderBookOrder(svcCtx *svc.ServiceContext, ctx context.Context, order *models.TTradeOrder) error {
	if svcCtx == nil || svcCtx.Redis == nil || order == nil {
		return nil
	}
	_, err := svcCtx.Redis.ZremCtx(ctx, orderBookKey(order), orderBookMember(order.Id))
	return err
}

func syncOrderBookCache(svcCtx *svc.ServiceContext, ctx context.Context, order *models.TTradeOrder) error {
	if isOrderBookOrder(order) {
		return cacheOrderBookOrder(svcCtx, ctx, order)
	}
	return removeOrderBookOrder(svcCtx, ctx, order)
}

func syncOrderBookCacheByID(svcCtx *svc.ServiceContext, ctx context.Context, orderID int64) error {
	if svcCtx == nil || orderID <= 0 {
		return nil
	}
	order, err := svcCtx.TradeOrderModel.FindOne(ctx, orderID)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return syncOrderBookCache(svcCtx, ctx, order)
}

func removeOrderBookMember(svcCtx *svc.ServiceContext, ctx context.Context, key string, orderID int64) error {
	if svcCtx == nil || svcCtx.Redis == nil || key == "" || orderID <= 0 {
		return nil
	}
	_, err := svcCtx.Redis.ZremCtx(ctx, key, orderBookMember(orderID))
	return err
}

func RestoreOrderBookCache(ctx context.Context, svcCtx *svc.ServiceContext) (int64, error) {
	if svcCtx == nil || svcCtx.TradeOrderModel == nil || svcCtx.Redis == nil {
		return 0, nil
	}

	var restored int64
	cursor := int64(0)
	for {
		orders, _, err := svcCtx.TradeOrderModel.FindPage(ctx, models.TradeOrderPageFilter{
			Statuses: matchableOrderStatuses(),
		}, cursor, orderBookRestoreSize)
		if err != nil {
			return restored, err
		}
		if len(orders) == 0 {
			return restored, nil
		}
		for _, order := range orders {
			cursor = order.Id
			if !isOrderBookOrder(order) {
				continue
			}
			if err := cacheOrderBookOrder(svcCtx, ctx, order); err != nil {
				return restored, err
			}
			restored++
		}
		if int64(len(orders)) < orderBookRestoreSize {
			return restored, nil
		}
	}
}
