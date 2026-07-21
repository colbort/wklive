package events

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/mq/kafka"
	"wklive/common/utils"
	"wklive/services/trade/internal/logic"
	"wklive/services/trade/internal/realtime"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		if err := svcCtx.TradeEventSubscriber.Subscribe(ctx, realtime.Channel, func(messageCtx context.Context, msg mq.Message) error {
			var event realtime.Event
			if err := mq.Decode(msg, &event); err != nil {
				logx.Errorf("decode trade real-time event failed: %v", err)
				return nil
			}
			if err := validateEvent(event); err != nil {
				logx.Errorf("invalid trade real-time event, eventNo=%s: %v", event.EventNo, err)
				_ = markEventFailed(svcCtx, messageCtx, event, err.Error())
				return nil
			}
			consumer := event.Consumer
			if consumer == "" {
				consumer = realtime.ConsumerTradeRealtime
			}
			now := utils.NowMillis()
			claimed, completed, lease, err := svcCtx.TradeEventInboxModel.Claim(messageCtx, consumer, event.TenantID, event.EventNo, event.Type, now, now-realtime.ClaimLeaseMillis)
			if err != nil {
				logx.Errorf("claim trade event inbox failed, eventNo=%s err=%v", event.EventNo, err)
				return nil
			}
			if completed {
				_ = markEventSuccess(svcCtx, messageCtx, event)
				return nil
			}
			if !claimed {
				return nil
			}
			if err := handleEvent(messageCtx, svcCtx, event); err != nil {
				logx.Errorf("handle trade real-time event failed, eventNo=%s type=%s bizId=%s err=%v", event.EventNo, event.Type, event.BizID, err)
				_ = svcCtx.TradeEventInboxModel.Fail(messageCtx, consumer, event.TenantID, event.EventNo, lease, err.Error(), utils.NowMillis())
				_ = markEventFailed(svcCtx, messageCtx, event, err.Error())
				return nil
			}
			completedLease, err := svcCtx.TradeEventInboxModel.Complete(messageCtx, consumer, event.TenantID, event.EventNo, lease, utils.NowMillis())
			if err != nil {
				logx.Errorf("complete trade event inbox failed, eventNo=%s err=%v", event.EventNo, err)
				return nil
			}
			if !completedLease {
				return nil
			}
			if err := markEventSuccess(svcCtx, messageCtx, event); err != nil {
				logx.Errorf("mark trade real-time event success failed, eventNo=%s err=%v", event.EventNo, err)
			}
			return nil
		}); err != nil && ctx.Err() == nil {
			logx.Errorf("trade real-time event subscriber stopped: %v", err)
		}
	}()
}

func validateEvent(event realtime.Event) error {
	if event.EventNo == "" || event.TenantID <= 0 || event.Type == "" {
		return errors.New("event_no, tenant_id and type are required")
	}
	if event.Version != realtime.PayloadVersionV1 {
		return fmt.Errorf("unsupported payload version %d", event.Version)
	}
	return nil
}

func handleEvent(ctx context.Context, svcCtx *svc.ServiceContext, event realtime.Event) error {
	switch event.Type {
	case realtime.EventOrderAccepted:
		orderID := event.OrderID
		if orderID <= 0 {
			order, err := svcCtx.TradeOrderModel.FindOneByTenantIdOrderNo(ctx, event.TenantID, event.BizID)
			if err != nil {
				return err
			}
			orderID = order.Id
		}
		return logic.NewProcessOrderMatchingLogic(ctx, svcCtx).ProcessOrder(orderID)
	case realtime.EventFillCreated:
		fillID := event.FillID
		if fillID <= 0 {
			fill, err := svcCtx.TradeFillModel.FindOneByTenantIdFillNo(ctx, event.TenantID, event.BizID)
			if err != nil {
				return err
			}
			fillID = fill.Id
		}
		return logic.NewProcessFillSettlementsLogic(ctx, svcCtx).ProcessFill(fillID)
	case realtime.EventPositionFill:
		fillID := event.FillID
		if fillID <= 0 {
			fill, err := svcCtx.TradeFillModel.FindOneByTenantIdFillNo(ctx, event.TenantID, event.BizID)
			if err != nil {
				return err
			}
			fillID = fill.Id
		}
		if err := logic.NewProcessContractPositionFillsLogic(ctx, svcCtx).ProcessFill(fillID); err != nil {
			return err
		}
		return logic.NewProcessFillSettlementsLogic(ctx, svcCtx).ProcessFill(fillID)
	default:
		// Domain notification events have no in-process side effect. Reaching
		// this consumer is their delivery acknowledgement; the full payload is
		// retained in both the Outbox row and the published message.
		return nil
	}
}

func findOutboxEvent(svcCtx *svc.ServiceContext, ctx context.Context, event realtime.Event) (*models.TBizTradeEvent, error) {
	if event.EventNo == "" {
		return nil, nil
	}
	item, err := svcCtx.BizTradeEventModel.FindOneByTenantIdEventNo(ctx, event.TenantID, event.EventNo)
	if errors.Is(err, models.ErrNotFound) {
		return nil, nil
	}
	return item, err
}

func markEventSuccess(svcCtx *svc.ServiceContext, ctx context.Context, event realtime.Event) error {
	item, err := findOutboxEvent(svcCtx, ctx, event)
	if err != nil || item == nil {
		return err
	}
	_, err = svcCtx.BizTradeEventModel.MarkDelivered(ctx, item.Id, event.ClaimToken, utils.NowMillis())
	return err
}

func markEventFailed(svcCtx *svc.ServiceContext, ctx context.Context, event realtime.Event, errorMessage string) error {
	item, err := findOutboxEvent(svcCtx, ctx, event)
	if err != nil || item == nil {
		return err
	}
	now := utils.NowMillis()
	delay := time.Duration(1) * time.Second
	retry := item.RetryCount
	if retry < 1 {
		retry = 1
	}
	if retry > 10 {
		retry = 10
	}
	delay *= time.Duration(1 << (retry - 1))
	_, err = svcCtx.BizTradeEventModel.MarkDeliveryFailed(ctx, item.Id, event.ClaimToken, now, now+delay.Milliseconds(), errorMessage)
	return err
}
