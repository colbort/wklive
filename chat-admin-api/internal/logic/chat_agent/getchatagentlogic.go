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

type GetChatAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatAgentLogic {
	return &GetChatAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChatAgentLogic) GetChatAgent(req *types.GetChatAgentReq) (resp *types.ChatAgentResp, err error) {
	return logicutil.Proxy[types.ChatAgentResp](l.ctx, req, l.svcCtx.ChatAdminCli.GetChatAgent)
}
