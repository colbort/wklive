// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_group

import (
	"context"

	"chat-admin-api/internal/logicutil"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatGroupsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageChatGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatGroupsLogic {
	return &PageChatGroupsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageChatGroupsLogic) PageChatGroups(req *types.PageChatGroupsReq) (resp *types.PageChatGroupsResp, err error) {
	return logicutil.Proxy[types.PageChatGroupsResp](l.ctx, req, l.svcCtx.ChatAdminCli.PageChatGroups)
}
