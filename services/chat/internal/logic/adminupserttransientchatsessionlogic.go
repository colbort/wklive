package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpsertTransientChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUpsertTransientChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpsertTransientChatSessionLogic {
	return &AdminUpsertTransientChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建或更新游客临时会话
func (l *AdminUpsertTransientChatSessionLogic) AdminUpsertTransientChatSession(in *chat.AdminUpsertTransientChatSessionReq) (*chat.AdminChatSessionResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatSessionResp{}, nil
}
