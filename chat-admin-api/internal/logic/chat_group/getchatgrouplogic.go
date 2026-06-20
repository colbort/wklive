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

type GetChatGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatGroupLogic {
	return &GetChatGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChatGroupLogic) GetChatGroup(req *types.GetChatGroupReq) (resp *types.ChatGroupResp, err error) {
	return logicutil.Proxy[types.ChatGroupResp](l.ctx, req, l.svcCtx.ChatAdminCli.GetChatGroup)
}
