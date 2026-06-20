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

type DeleteChatGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatGroupLogic {
	return &DeleteChatGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteChatGroupLogic) DeleteChatGroup(req *types.DeleteChatGroupReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.ChatAdminCli.DeleteChatGroup)
}
