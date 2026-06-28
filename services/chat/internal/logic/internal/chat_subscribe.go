package internal

import (
	"context"
	"fmt"

	v9 "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/protobuf/encoding/protojson"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
)

type ChatEventStream interface {
	Context() context.Context
	Send(*chat.ChatMessageEvent) error
}

type ChatSubscribeRequest interface {
	GetMerchantId() int64
	GetUserId() int64
	GetAgentId() int64
	GetSessionNo() string
	GetIsGuest() bool
	GetAdmin() bool
}

type SubscribeOptions struct {
	Channel string
	Admin   bool
}

type redisSubscriber interface {
	Subscribe(ctx context.Context, channels ...string) *v9.PubSub
}

func SubscribeChatEventStream(svcCtx *svc.ServiceContext, req ChatSubscribeRequest, stream ChatEventStream, opts SubscribeOptions) error {
	if svcCtx == nil || svcCtx.BusRedis == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	rds, err := redis.NewRedis(svcCtx.Config.Redis.RedisConf)
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

	ctx := stream.Context()
	pubsub := client.Subscribe(ctx, opts.Channel)
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
			var event chat.ChatMessageEvent
			if err := protojson.Unmarshal([]byte(msg.Payload), &event); err != nil {
				logx.WithContext(ctx).Errorf("decode chat stream event failed: %v", err)
				continue
			}
			if err := stream.Send(&event); err != nil {
				return err
			}
		}
	}
}
