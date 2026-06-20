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

type ListEnabledChatCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListEnabledChatCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEnabledChatCategoriesLogic {
	return &ListEnabledChatCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListEnabledChatCategoriesLogic) ListEnabledChatCategories(req *types.ListEnabledChatCategoriesReq) (resp *types.ListChatCategoriesResp, err error) {
	return logicutil.Proxy[types.ListChatCategoriesResp](l.ctx, req, l.svcCtx.ChatAdminCli.ListEnabledChatCategories)
}
