package ws

import (
	"context"
	"time"

	"wklive/common/mq/kafka"
	"wklive/common/notify"

	"github.com/zeromicro/go-zero/core/logx"
)

func SubscribeMQ(ctx context.Context, subscriber *mq.Subscriber, hub *Hub) {
	for {
		err := subscriber.Subscribe(ctx, notify.Channel, func(_ context.Context, message mq.Message) error {
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
