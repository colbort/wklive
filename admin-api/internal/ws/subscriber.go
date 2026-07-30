package ws

import (
	"context"
	"encoding/json"
	"time"

	"wklive/common/mq/kafka"
	"wklive/common/notify"

	"github.com/zeromicro/go-zero/core/logx"
)

func SubscribeMQ(
	ctx context.Context,
	subscriber *mq.Subscriber,
	hub *Hub,
	store *NotificationStore,
) {
	for {
		err := subscriber.Subscribe(ctx, notify.Channel, func(messageCtx context.Context, message mq.Message) error {
			if store != nil {
				var event notify.Event
				if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
					return err
				}
				if err := store.RecordEvent(messageCtx, event); err != nil {
					return err
				}
			}
			hub.BroadcastRaw([]byte(message.Payload))
			return nil
		})
		if ctx.Err() != nil {
			return
		}
		logx.Errorf("admin notification mq subscription failed: %v", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}
