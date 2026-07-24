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
		logx.WithContext(ctx).Error("seconds delay queue consumer not started: delay queue is disabled")
		return
	}
	logx.WithContext(ctx).Info("seconds delay queue consumer starting")
	go svcCtx.DelayQueue.Consume(func(message delayqueue.Message) {
		if message.Action == delayqueue.ActionActivate || message.Action == delayqueue.ActionSettle {
			logx.WithContext(ctx).Infof(
				"seconds lifecycle stage=message_received action=%s tenantId=%d orderId=%d version=%d dueAt=%d",
				message.Action, message.TenantID, message.OrderID, message.Version, message.DueAt,
			)
		}
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
			logx.WithContext(ctx).Infof(
				"seconds lifecycle stage=activation_stale tenantId=%d orderId=%d messageVersion=%d currentVersion=%d currentTenantId=%d",
				message.TenantID, message.OrderID, message.Version, order.Version, order.TenantId,
			)
			return nil
		}
		processErr := NewProcessSecondsSettlementsLogic(ctx, svcCtx).Process(message.TenantID)
		item, err := svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(ctx, message.TenantID, message.OrderID)
		if err != nil {
			return errors.Join(processErr, err)
		}
		if item.SettlementStatus != int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVE) {
			logx.WithContext(ctx).Errorf(
				"seconds lifecycle stage=activation_not_active tenantId=%d orderId=%d settlementStatus=%d retryCount=%d nextRetryAt=%d lastError=%q err=%v",
				message.TenantID, message.OrderID, item.SettlementStatus, item.RetryCount, item.NextRetryAt, item.LastErrorMsg, processErr,
			)
			return processErr
		}
		enqueueErr := svcCtx.DelayQueue.At(delayqueue.Message{
			Action:   delayqueue.ActionSettle,
			TenantID: message.TenantID,
			OrderID:  message.OrderID,
			Version:  item.Version,
			DueAt:    item.ExpireTime,
		}, time.UnixMilli(item.ExpireTime))
		if enqueueErr == nil {
			logx.WithContext(ctx).Infof(
				"seconds lifecycle stage=settlement_enqueued tenantId=%d orderId=%d version=%d expireTime=%d",
				message.TenantID, message.OrderID, item.Version, item.ExpireTime,
			)
		}
		return errors.Join(processErr, enqueueErr)
	case delayqueue.ActionSettle:
		if message.TenantID <= 0 || message.OrderID <= 0 {
			return errors.New("invalid seconds settlement message")
		}
		if now := time.Now().UnixMilli(); message.DueAt > now {
			logx.WithContext(ctx).Infof(
				"seconds lifecycle stage=settlement_early_requeued tenantId=%d orderId=%d dueAt=%d now=%d",
				message.TenantID, message.OrderID, message.DueAt, now,
			)
			return svcCtx.DelayQueue.At(message, time.UnixMilli(message.DueAt))
		}
		item, err := svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(ctx, message.TenantID, message.OrderID)
		if err != nil {
			return err
		}
		if item.Version != message.Version || item.ExpireTime != message.DueAt {
			logx.WithContext(ctx).Infof(
				"seconds lifecycle stage=settlement_stale tenantId=%d orderId=%d messageVersion=%d currentVersion=%d messageDueAt=%d currentExpireTime=%d",
				message.TenantID, message.OrderID, message.Version, item.Version, message.DueAt, item.ExpireTime,
			)
			return nil
		}
		processErr := NewProcessSecondsSettlementsLogic(ctx, svcCtx).Process(message.TenantID)
		current, findErr := svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(ctx, message.TenantID, message.OrderID)
		if findErr == nil {
			logx.WithContext(ctx).Infof(
				"seconds lifecycle stage=settlement_processed tenantId=%d orderId=%d settlementStatus=%d result=%d retryCount=%d lastError=%q err=%v",
				message.TenantID, message.OrderID, current.SettlementStatus, current.Result, current.RetryCount, current.LastErrorMsg, processErr,
			)
			if processErr != nil &&
				current.SettlementStatus == int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING) &&
				current.NextRetryAt > time.Now().UnixMilli() {
				retryMessage := delayqueue.Message{
					Action:   delayqueue.ActionSettle,
					TenantID: current.TenantId,
					OrderID:  current.OrderId,
					Version:  current.Version,
					DueAt:    current.ExpireTime,
				}
				retryErr := svcCtx.DelayQueue.At(retryMessage, time.UnixMilli(current.NextRetryAt))
				if retryErr == nil {
					logx.WithContext(ctx).Infof(
						"seconds lifecycle stage=settlement_retry_enqueued tenantId=%d orderId=%d version=%d retryCount=%d nextRetryAt=%d expireTime=%d",
						current.TenantId, current.OrderId, current.Version, current.RetryCount, current.NextRetryAt, current.ExpireTime,
					)
				}
				processErr = errors.Join(processErr, retryErr)
			}
		}
		return errors.Join(processErr, findErr)
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
