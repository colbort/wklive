package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/proto/trade"
	"wklive/services/trade/internal/delayqueue"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func enqueueSecondsActivation(svcCtx *svc.ServiceContext, order *models.TTradeOrder) error {
	if svcCtx == nil || svcCtx.DelayQueue == nil || order == nil {
		return nil
	}
	return svcCtx.DelayQueue.Delay(delayqueue.Message{
		Action:   delayqueue.ActionActivate,
		TenantID: order.TenantId,
		OrderID:  order.Id,
		Version:  order.Version,
	}, time.Millisecond)
}

func enqueueOrderExpiration(svcCtx *svc.ServiceContext, order *models.TTradeOrder) error {
	if svcCtx == nil || svcCtx.DelayQueue == nil || order == nil || order.ExpireAt <= 0 {
		return nil
	}
	return svcCtx.DelayQueue.At(delayqueue.Message{
		Action: delayqueue.ActionExpireOrder, TenantID: order.TenantId, OrderID: order.Id,
		Version: order.Version, DueAt: order.ExpireAt,
	}, time.UnixMilli(order.ExpireAt))
}

func StartSecondsDelayQueue(ctx context.Context, svcCtx *svc.ServiceContext) {
	if svcCtx == nil || svcCtx.DelayQueue == nil {
		return
	}
	go svcCtx.DelayQueue.Consume(func(message delayqueue.Message) {
		if err := handleSecondsDelayMessage(ctx, svcCtx, message); err != nil {
			logx.WithContext(ctx).Errorf(
				"seconds delay message failed, action=%s tenantId=%d orderId=%d version=%d err=%v",
				message.Action, message.TenantID, message.OrderID, message.Version, err,
			)
		}
	})
}

func handleSecondsDelayMessage(ctx context.Context, svcCtx *svc.ServiceContext, message delayqueue.Message) error {
	switch message.Action {
	case delayqueue.ActionActivate:
		if message.TenantID <= 0 || message.OrderID <= 0 {
			return errors.New("invalid seconds activation message")
		}
		order, err := svcCtx.TradeOrderModel.FindOne(ctx, message.OrderID)
		if err != nil {
			return err
		}
		if order.TenantId != message.TenantID || order.Version != message.Version {
			return nil
		}
		processErr := NewProcessSecondsSettlementsLogic(ctx, svcCtx).Process(message.TenantID)
		item, err := svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(ctx, message.TenantID, message.OrderID)
		if err != nil {
			return errors.Join(processErr, err)
		}
		if item.SettlementStatus != int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVE) {
			return processErr
		}
		enqueueErr := svcCtx.DelayQueue.At(delayqueue.Message{
			Action:   delayqueue.ActionSettle,
			TenantID: message.TenantID,
			OrderID:  message.OrderID,
			Version:  item.Version,
			DueAt:    item.ExpireTime,
		}, time.UnixMilli(item.ExpireTime))
		return errors.Join(processErr, enqueueErr)
	case delayqueue.ActionSettle:
		if message.TenantID <= 0 || message.OrderID <= 0 {
			return errors.New("invalid seconds settlement message")
		}
		item, err := svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(ctx, message.TenantID, message.OrderID)
		if err != nil {
			return err
		}
		if item.Version != message.Version || item.ExpireTime != message.DueAt {
			return nil
		}
		return NewProcessSecondsSettlementsLogic(ctx, svcCtx).Process(message.TenantID)
	case delayqueue.ActionExpireOrder:
		if message.TenantID <= 0 || message.OrderID <= 0 {
			return errors.New("invalid order expiration message")
		}
		order, err := svcCtx.TradeOrderModel.FindOne(ctx, message.OrderID)
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		// Matching and partial fills legitimately advance the order version.
		// ExpireAt identifies a stale rescheduled message without suppressing
		// expiration after an ordinary fill.
		if order.TenantId != message.TenantID || order.ExpireAt != message.DueAt {
			return nil
		}
		expired, err := NewProcessTradeEventsLogic(ctx, svcCtx).expireOrderIfNeeded(order.Id, time.Now().UnixMilli())
		if err != nil || expired == nil {
			return err
		}
		if err = unfreezeRemainingOrderAsset(svcCtx, ctx, expired, "trade delayed expiration unfreeze"); err != nil {
			return err
		}
		if err = removeOrderBookOrder(svcCtx, ctx, expired); err != nil {
			logx.WithContext(ctx).Errorf("remove delayed expired order from cache failed, orderId=%d err=%v", expired.Id, err)
		}
		return nil
	case delayqueue.ActionExpireTradeRisk:
		return expireTradeRiskMessage(ctx, svcCtx, message)
	case delayqueue.ActionExpireSymbolRisk:
		return expireSymbolRiskMessage(ctx, svcCtx, message)
	default:
		return fmt.Errorf("unsupported seconds delay action: %s", message.Action)
	}
}

func expireTradeRiskMessage(ctx context.Context, svcCtx *svc.ServiceContext, message delayqueue.Message) error {
	item, err := svcCtx.RiskUserTradeLimitModel.FindOne(ctx, message.EntityID)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if item.TenantId != message.TenantID || item.EffectiveEndTime != message.DueAt ||
		item.EffectiveEndTime > time.Now().UnixMilli() {
		return nil
	}
	item.Enabled = 2
	item.UpdateTimes = time.Now().UnixMilli()
	return svcCtx.RiskUserTradeLimitModel.Update(ctx, item)
}

func expireSymbolRiskMessage(ctx context.Context, svcCtx *svc.ServiceContext, message delayqueue.Message) error {
	item, err := svcCtx.RiskUserSymbolLimitModel.FindOne(ctx, message.EntityID)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if item.TenantId != message.TenantID || item.EffectiveEndTime != message.DueAt ||
		item.EffectiveEndTime > time.Now().UnixMilli() {
		return nil
	}
	item.Enabled = 2
	item.UpdateTimes = time.Now().UnixMilli()
	return svcCtx.RiskUserSymbolLimitModel.Update(ctx, item)
}
