package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminSubscribeStreamLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminSubscribeStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSubscribeStreamLogic {
	return &AdminSubscribeStreamLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订阅客服消息事件流
func (l *AdminSubscribeStreamLogic) AdminSubscribeStream(in *chat.AdminChatSubscribeRequest, stream chat.ChatAdmin_AdminSubscribeStreamServer) error {
	// todo: add your logic here and delete this line

	return nil
}
