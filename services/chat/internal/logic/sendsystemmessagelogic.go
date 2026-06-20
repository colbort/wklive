package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendSystemMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendSystemMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSystemMessageLogic {
	return &SendSystemMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送系统消息
func (l *SendSystemMessageLogic) SendSystemMessage(in *chat.SendSystemMessageReq) (*chat.InternalChatMessageResp, error) {
	// todo: add your logic here and delete this line

	return &chat.InternalChatMessageResp{}, nil
}
