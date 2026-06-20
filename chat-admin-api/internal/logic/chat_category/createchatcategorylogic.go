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

type CreateChatCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateChatCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatCategoryLogic {
	return &CreateChatCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChatCategoryLogic) CreateChatCategory(req *types.CreateChatCategoryReq) (resp *types.ChatCategoryResp, err error) {
	return logicutil.Proxy[types.ChatCategoryResp](l.ctx, req, l.svcCtx.ChatAdminCli.CreateChatCategory)
}
