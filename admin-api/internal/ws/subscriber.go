package ws

import (
	"context"
	"fmt"
	"time"

	"wklive/common/notify"

	v9 "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func SubscribeRedis(ctx context.Context, conf redis.RedisConf, hub *Hub) {
	for {
		if err := subscribeRedis(ctx, conf, hub); err != nil {
			logx.Errorf("admin notification subscribe failed: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func subscribeRedis(ctx context.Context, conf redis.RedisConf, hub *Hub) error {
	rds, err := redis.NewRedis(conf)
	if err != nil {
		return err
	}

	node, err := redis.CreateBlockingNode(rds)
	if err != nil {
		return err
	}
	defer node.Close()

	client, ok := node.(redisSubscriber)
	if !ok {
		return fmt.Errorf("redis node does not support subscribe")
	}

	pubsub := client.Subscribe(ctx, notify.Channel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			hub.BroadcastRaw([]byte(msg.Payload))
		}
	}
}

type redisSubscriber interface {
	Subscribe(ctx context.Context, channels ...string) *v9.PubSub
}
