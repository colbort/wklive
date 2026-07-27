package realtime

import (
	"context"
	"fmt"
	"time"

	mq "wklive/common/mq/kafka"
	"wklive/common/userevent"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartUserEventSubscriber(ctx context.Context, subscriber *mq.Subscriber, hub *UserEventHub) {
	go func() {
		for ctx.Err() == nil {
			err := subscriber.Subscribe(ctx, userevent.Channel, func(_ context.Context, message mq.Message) error {
				var event UserEvent
				if err := mq.Decode(message, &event); err != nil {
					return fmt.Errorf("decode private user event: %w", err)
				}
				hub.Publish(event)
				return nil
			})
			if ctx.Err() != nil {
				return
			}
			logx.Errorf("private user event subscriber stopped: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}()
}
