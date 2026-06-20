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

type GetChatCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChatCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatCategoryLogic {
	return &GetChatCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChatCategoryLogic) GetChatCategory(req *types.GetChatCategoryReq) (resp *types.ChatCategoryResp, err error) {
	return logicutil.Proxy[types.ChatCategoryResp](l.ctx, req, l.svcCtx.ChatAdminCli.GetChatCategory)
}
