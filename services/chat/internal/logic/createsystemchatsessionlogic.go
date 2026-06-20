package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSystemChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSystemChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSystemChatSessionLogic {
	return &CreateSystemChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建系统会话
func (l *CreateSystemChatSessionLogic) CreateSystemChatSession(in *chat.CreateSystemChatSessionReq) (*chat.InternalChatSessionResp, error) {
	// todo: add your logic here and delete this line

	return &chat.InternalChatSessionResp{}, nil
}
