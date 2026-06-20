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
	data, base, err := assignSession(l.ctx, l.svcCtx, in)
	if err != nil {
		return &chat.AdminChatSessionResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AdminChatSessionResp{Base: base}, nil
	}
	return &chat.AdminChatSessionResp{Base: okBase(), Data: toProtoSession(data)}, nil
}
