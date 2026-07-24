package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/proto/trade"
	"wklive/services/trade/internal/secondsqueue"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func enqueueSecondsActivation(svcCtx *svc.ServiceContext, order *models.TTradeOrder) error {
	if svcCtx == nil || svcCtx.SecondsDelayQueue == nil || order == nil {
		return nil
	}
	return svcCtx.SecondsDelayQueue.Delay(secondsqueue.Message{
		Action:   secondsqueue.ActionActivate,
		TenantID: order.TenantId,
		OrderID:  order.Id,
		Version:  order.Version,
	}, time.Millisecond)
}

func StartSecondsDelayQueue(ctx context.Context, svcCtx *svc.ServiceContext) {
	if svcCtx == nil || svcCtx.SecondsDelayQueue == nil {
		return
	}
	go svcCtx.SecondsDelayQueue.Consume(func(message secondsqueue.Message) {
		if err := handleSecondsDelayMessage(ctx, svcCtx, message); err != nil {
			logx.WithContext(ctx).Errorf(
				"seconds delay message failed, action=%s tenantId=%d orderId=%d version=%d err=%v",
				message.Action, message.TenantID, message.OrderID, message.Version, err,
			)
		}
	})
}

func handleSecondsDelayMessage(ctx context.Context, svcCtx *svc.ServiceContext, message secondsqueue.Message) error {
	if message.TenantID <= 0 || message.OrderID <= 0 {
		return errors.New("invalid seconds delay message")
	}
	switch message.Action {
	case secondsqueue.ActionActivate:
		processErr := NewProcessSecondsSettlementsLogic(ctx, svcCtx).Process(message.TenantID)
		item, err := svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(ctx, message.TenantID, message.OrderID)
		if err != nil {
			return errors.Join(processErr, err)
		}
		if item.SettlementStatus != int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVE) {
			return processErr
		}
		enqueueErr := svcCtx.SecondsDelayQueue.At(secondsqueue.Message{
			Action:   secondsqueue.ActionSettle,
			TenantID: message.TenantID,
			OrderID:  message.OrderID,
			Version:  item.Version,
		}, time.UnixMilli(item.ExpireTime))
		return errors.Join(processErr, enqueueErr)
	case secondsqueue.ActionSettle:
		return NewProcessSecondsSettlementsLogic(ctx, svcCtx).Process(message.TenantID)
	default:
		return fmt.Errorf("unsupported seconds delay action: %s", message.Action)
	}
}
