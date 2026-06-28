package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppSubscribeStreamLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppSubscribeStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppSubscribeStreamLogic {
	return &AppSubscribeStreamLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订阅客服消息事件流
func (l *AppSubscribeStreamLogic) AppSubscribeStream(in *chat.AppChatSubscribeRequest, stream chat.ChatApp_AppSubscribeStreamServer) error {
	return internal.SubscribeChatEventStream(l.svcCtx, in, stream, internal.SubscribeOptions{
		Channel: chat.ChatAppEventChannel,
		Admin:   false,
	})
}
