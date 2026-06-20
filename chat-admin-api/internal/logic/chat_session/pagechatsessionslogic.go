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

type PageChatSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageChatSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatSessionsLogic {
	return &PageChatSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageChatSessionsLogic) PageChatSessions(req *types.PageChatSessionsReq) (resp *types.PageChatSessionsResp, err error) {
	return logicutil.Proxy[types.PageChatSessionsResp](l.ctx, req, l.svcCtx.ChatAdminCli.PageChatSessions)
}
