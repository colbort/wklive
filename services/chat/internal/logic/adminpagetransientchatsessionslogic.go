package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminPageTransientChatSessionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminPageTransientChatSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminPageTransientChatSessionsLogic {
	return &AdminPageTransientChatSessionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询游客临时会话
func (l *AdminPageTransientChatSessionsLogic) AdminPageTransientChatSessions(in *chat.AdminPageTransientChatSessionsReq) (*chat.AdminPageTransientChatSessionsResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminPageTransientChatSessionsResp{}, nil
}
