package helper

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
)

type ChatEventStream interface {
	Context() context.Context
	Send(*chat.ChatWsResponse) error
}

type SubscribeOptions struct {
	Channel string
	Admin   bool
}

func SubscribeChatEventStream(svcCtx *svc.ServiceContext, stream ChatEventStream, opts SubscribeOptions) error {
	ctx := stream.Context()
	ch, unsubscribe := svcCtx.ChatEventHub.Subscribe(opts.Channel)
	defer unsubscribe()
	for {
		select {
		case <-ctx.Done():
			return nil
		case payload, ok := <-ch:
			if !ok {
				return nil
			}
			var event chat.ChatWsResponse
			if err := protojson.Unmarshal(payload, &event); err != nil {
				logx.WithContext(ctx).Errorf("decode chat stream event failed: %v", err)
				continue
			}
			if err := stream.Send(&event); err != nil {
				return err
			}
		}
	}
}
