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

type UpdateChatCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateChatCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatCategoryLogic {
	return &UpdateChatCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateChatCategoryLogic) UpdateChatCategory(req *types.UpdateChatCategoryReq) (resp *types.ChatCategoryResp, err error) {
	return logicutil.Proxy[types.ChatCategoryResp](l.ctx, req, l.svcCtx.ChatAdminCli.UpdateChatCategory)
}
