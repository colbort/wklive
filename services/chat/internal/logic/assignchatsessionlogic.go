package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignChatSessionLogic {
	return &AssignChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分配会话
func (l *AssignChatSessionLogic) AssignChatSession(in *chat.AssignChatSessionReq) (*chat.AdminChatSessionResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatSessionResp{}, nil
}
