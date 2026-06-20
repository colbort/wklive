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

type UpdateChatAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatAgentLogic {
	return &UpdateChatAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateChatAgentLogic) UpdateChatAgent(req *types.UpdateChatAgentReq) (resp *types.ChatAgentResp, err error) {
	return logicutil.Proxy[types.ChatAgentResp](l.ctx, req, l.svcCtx.ChatAdminCli.UpdateChatAgent)
}
