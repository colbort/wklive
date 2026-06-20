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

type UpdateChatQuickReplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateChatQuickReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatQuickReplyLogic {
	return &UpdateChatQuickReplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateChatQuickReplyLogic) UpdateChatQuickReply(req *types.UpdateChatQuickReplyReq) (resp *types.ChatQuickReplyResp, err error) {
	return logicutil.Proxy[types.ChatQuickReplyResp](l.ctx, req, l.svcCtx.ChatAdminCli.UpdateChatQuickReply)
}
