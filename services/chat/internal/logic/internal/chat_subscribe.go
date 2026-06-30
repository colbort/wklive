package internal

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
)

type ChatEventStream interface {
	Context() context.Context
	Send(*chat.ChatMessageEvent) error
}

type SubscribeOptions struct {
	Channel string
	Admin   bool
}

func SubscribeChatEventStream(svcCtx *svc.ServiceContext, stream ChatEventStream, opts SubscribeOptions) error {
	busRedis := redis.NewClient(&redis.Options{
		Addr:     svcCtx.Config.CacheRedis[0].Host,
		Username: svcCtx.Config.CacheRedis[0].User,
		Password: svcCtx.Config.CacheRedis[0].Pass,
		DB:       0,
	})
	ctx := stream.Context()
	pubsub := busRedis.Subscribe(ctx, opts.Channel)
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
