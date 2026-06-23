// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_session

import (
	"context"

	"chat-admin-api/internal/logicutil"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignChatSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignChatSessionLogic {
	return &AssignChatSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignChatSessionLogic) AssignChatSession(req *types.AssignChatSessionReq) (resp *types.ChatSessionResp, err error) {
	resp, err = logicutil.Proxy[types.ChatSessionResp](l.ctx, req, l.svcCtx.ChatAdminCli.AssignChatSession)
	if resp != nil {
		enrichSession(&resp.Data)
	}
	return resp, err
}
