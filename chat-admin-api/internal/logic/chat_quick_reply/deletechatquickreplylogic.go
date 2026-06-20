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

type DeleteChatQuickReplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteChatQuickReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatQuickReplyLogic {
	return &DeleteChatQuickReplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteChatQuickReplyLogic) DeleteChatQuickReply(req *types.DeleteChatQuickReplyReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.ChatAdminCli.DeleteChatQuickReply)
}
