// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_category

import (
	"context"

	"chat-admin-api/internal/logicutil"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteChatCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteChatCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatCategoryLogic {
	return &DeleteChatCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteChatCategoryLogic) DeleteChatCategory(req *types.DeleteChatCategoryReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.ChatAdminCli.DeleteChatCategory)
}
