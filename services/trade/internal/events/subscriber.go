package events

import (
	"context"
	"errors"

	bus "wklive/common/bus/redis"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/logic"
	"wklive/services/trade/internal/realtime"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		if err := svcCtx.TradeEventSubscriber.Subscribe(ctx, realtime.Channel, func(messageCtx context.Context, msg bus.Message) error {
			var event realtime.Event
			if err := bus.Decode(msg, &event); err != nil {
				logx.Errorf("decode trade real-time event failed: %v", err)
				return nil
			}
			if err := handleEvent(messageCtx, svcCtx, event); err != nil {
				logx.Errorf("handle trade real-time event failed, eventNo=%s type=%s bizId=%s err=%v", event.EventNo, event.Type, event.BizID, err)
				_ = markEvent(svcCtx, messageCtx, event, false, err.Error())
				return nil
			}
			if err := markEvent(svcCtx, messageCtx, event, true, ""); err != nil {
				logx.Errorf("mark trade real-time event success failed, eventNo=%s err=%v", event.EventNo, err)
			}
			return nil
		}); err != nil && ctx.Err() == nil {
			logx.Errorf("trade real-time event subscriber stopped: %v", err)
		}
	}()
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
		return logic.NewProcessSpotSettlementsLogic(ctx, svcCtx).ProcessFill(fillID)
	default:
		return nil
	}
}

func markEvent(svcCtx *svc.ServiceContext, ctx context.Context, event realtime.Event, success bool, errorMessage string) error {
	if event.EventNo == "" {
		return nil
	}
	item, err := svcCtx.BizTradeEventModel.FindOneByTenantIdEventNo(ctx, event.TenantID, event.EventNo)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if success {
		item.EventStatus = int64(trade.EventStatus_EVENT_STATUS_SUCCESS)
		item.LastErrorMsg = ""
		item.NextRetryAt = 0
	} else {
		item.EventStatus = int64(trade.EventStatus_EVENT_STATUS_FAILED)
		item.RetryCount++
		item.LastErrorMsg = errorMessage
		item.NextRetryAt = utils.NowMillis() + 1000
	}
	item.UpdateTimes = utils.NowMillis()
	return svcCtx.BizTradeEventModel.Update(ctx, item)
}
