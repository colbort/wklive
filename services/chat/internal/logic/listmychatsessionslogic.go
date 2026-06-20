package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyChatSessionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyChatSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyChatSessionsLogic {
	return &ListMyChatSessionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的会话列表
func (l *ListMyChatSessionsLogic) ListMyChatSessions(in *chat.ListMyChatSessionsReq) (*chat.ListChatSessionsResp, error) {
	// todo: add your logic here and delete this line

	return &chat.ListChatSessionsResp{}, nil
}
