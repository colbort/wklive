// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_work_order

import (
	"context"

	"chat-admin-api/internal/logicutil"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatWorkOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatWorkOrderLogic {
	return &UpdateChatWorkOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateChatWorkOrderLogic) UpdateChatWorkOrder(req *types.UpdateChatWorkOrderReq) (resp *types.ChatWorkOrderResp, err error) {
	return logicutil.Proxy[types.ChatWorkOrderResp](l.ctx, req, l.svcCtx.ChatAdminCli.UpdateChatWorkOrder)
}
