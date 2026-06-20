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

type MarkAgentMessagesReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkAgentMessagesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAgentMessagesReadLogic {
	return &MarkAgentMessagesReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkAgentMessagesReadLogic) MarkAgentMessagesRead(req *types.MarkAgentMessagesReadReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.ChatAdminCli.MarkAgentMessagesRead)
}
