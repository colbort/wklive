// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_quick_reply

import (
	"context"

	"chat-admin-api/internal/logicutil"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatQuickRepliesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageChatQuickRepliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatQuickRepliesLogic {
	return &PageChatQuickRepliesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageChatQuickRepliesLogic) PageChatQuickReplies(req *types.PageChatQuickRepliesReq) (resp *types.ListChatQuickRepliesResp, err error) {
	return logicutil.Proxy[types.ListChatQuickRepliesResp](l.ctx, req, l.svcCtx.ChatAdminCli.PageChatQuickReplies)
}
