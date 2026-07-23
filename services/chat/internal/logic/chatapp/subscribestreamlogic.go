package chatapplogic

import (
	"context"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubscribeStreamLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubscribeStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeStreamLogic {
	return &SubscribeStreamLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订阅客服消息事件流
func (l *SubscribeStreamLogic) SubscribeStream(in *chat.SubscribeRequest, stream chat.ChatApp_SubscribeStreamServer) error {
	return ih.SubscribeChatEventStream(l.svcCtx, stream, ih.SubscribeOptions{
		Channel: chat.ChatAppEventChannel,
		Admin:   false,
	})
}
