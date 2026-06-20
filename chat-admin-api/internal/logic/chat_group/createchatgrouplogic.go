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

type CreateChatGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatGroupLogic {
	return &CreateChatGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChatGroupLogic) CreateChatGroup(req *types.CreateChatGroupReq) (resp *types.ChatGroupResp, err error) {
	return logicutil.Proxy[types.ChatGroupResp](l.ctx, req, l.svcCtx.ChatAdminCli.CreateChatGroup)
}
