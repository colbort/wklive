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

type CreateChatAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatAgentLogic {
	return &CreateChatAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChatAgentLogic) CreateChatAgent(req *types.CreateChatAgentReq) (resp *types.ChatAgentResp, err error) {
	return logicutil.Proxy[types.ChatAgentResp](l.ctx, req, l.svcCtx.ChatAdminCli.CreateChatAgent)
}
