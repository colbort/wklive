package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetTransientChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetTransientChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetTransientChatSessionLogic {
	return &AdminGetTransientChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询游客临时会话
func (l *AdminGetTransientChatSessionLogic) AdminGetTransientChatSession(in *chat.AdminGetTransientChatSessionReq) (*chat.AdminChatSessionResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatSessionResp{}, nil
}
