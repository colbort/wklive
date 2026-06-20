// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_agent

import (
	"context"

	"chat-admin-api/internal/logicutil"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageChatAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatAgentsLogic {
	return &PageChatAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageChatAgentsLogic) PageChatAgents(req *types.PageChatAgentsReq) (resp *types.PageChatAgentsResp, err error) {
	return logicutil.Proxy[types.PageChatAgentsResp](l.ctx, req, l.svcCtx.ChatAdminCli.PageChatAgents)
}
