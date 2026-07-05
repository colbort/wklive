// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_auth

import (
	"context"

	"chat-admin-api/internal/logicutil"
	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateChatConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatConfigLogic {
	return &UpdateChatConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateChatConfigLogic) UpdateChatConfig(req *types.UpdateChatConfigReq) (resp *types.ChatConfigResp, err error) {
	return logicutil.Proxy[types.ChatConfigResp](l.ctx, req, l.svcCtx.ChatAdminCli.UpdateChatConfig)
}
