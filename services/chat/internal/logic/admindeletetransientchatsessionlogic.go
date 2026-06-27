package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminDeleteTransientChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminDeleteTransientChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeleteTransientChatSessionLogic {
	return &AdminDeleteTransientChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除游客临时会话和消息
func (l *AdminDeleteTransientChatSessionLogic) AdminDeleteTransientChatSession(in *chat.AdminDeleteTransientChatSessionReq) (*chat.AdminDeleteTransientChatSessionResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminDeleteTransientChatSessionResp{}, nil
}
