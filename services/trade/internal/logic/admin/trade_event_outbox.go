package adminlogic

import (
	"context"
	"errors"
	"time"

	"wklive/common/utils"
	"wklive/services/trade/internal/realtime"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"
)

const tradeEventPayloadVersion = realtime.PayloadVersionV1

func tradeEventConsumer(eventType string) string {
	switch eventType {
	case realtime.EventOrderAccepted, realtime.EventFillCreated, realtime.EventPositionFill:
		return realtime.ConsumerTradeRealtime
	default:
		return "trade-domain"
	}
}

func publishTradeOutboxEvent(ctx context.Context, svcCtx *svc.ServiceContext, event realtime.Event) error {
	item, err := svcCtx.BizTradeEventModel.FindOneByTenantIdEventNo(ctx, event.TenantID, event.EventNo)
	if errors.Is(err, models.ErrNotFound) {
		return errors.New("trade outbox event not found")
	}
	if err != nil {
		return err
	}
	now := utils.NowMillis()
	if item.EventStatus == 5 && item.ClaimedAt <= now-realtime.ClaimLeaseMillis && item.MaxRetryCount > 0 && item.RetryCount >= item.MaxRetryCount {
		_, err = svcCtx.BizTradeEventModel.MarkDeliveryFailed(ctx, item.Id, item.ClaimedBy, now, 0, "delivery acknowledgement lease expired after maximum attempts")
		return err
	}
	claimed, err := svcCtx.BizTradeEventModel.ClaimDispatch(ctx, item.Id, svcCtx.TradeEventInstanceID, now, now-realtime.ClaimLeaseMillis)
	if err != nil || !claimed {
		return err
	}
	if event.Version == 0 {
		event.Version = item.PayloadVersion
	}
	if event.Consumer == "" {
		event.Consumer = item.Consumer
	}
	if event.Payload == "" {
		event.Payload = item.Payload
	}
	event.ClaimToken = svcCtx.TradeEventInstanceID
	if event.Version == 0 {
		event.Version = tradeEventPayloadVersion
	}
	if err := realtime.Publish(ctx, svcCtx.TradeEventPublisher, event); err != nil {
		_, markErr := svcCtx.BizTradeEventModel.MarkDeliveryFailed(ctx, item.Id, svcCtx.TradeEventInstanceID, now, now+tradeEventRetryDelay(item.RetryCount+1).Milliseconds(), err.Error())
		if markErr != nil {
			return markErr
		}
		return err
	}
	return nil
}

func tradeEventRetryDelay(retryCount int64) time.Duration {
	if retryCount < 1 {
		retryCount = 1
	}
	if retryCount > 10 {
		retryCount = 10
	}
	return time.Second * time.Duration(1<<(retryCount-1))
}
