package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyChatMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyChatMessagesLogic {
	return &ListMyChatMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询会话消息
func (l *ListMyChatMessagesLogic) ListMyChatMessages(in *chat.ListMyChatMessagesReq) (*chat.ListChatMessagesResp, error) {
	// todo: add your logic here and delete this line

	return &chat.ListChatMessagesResp{}, nil
}
