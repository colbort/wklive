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

type CreateChatWorkOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatWorkOrderLogic {
	return &CreateChatWorkOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChatWorkOrderLogic) CreateChatWorkOrder(req *types.CreateChatWorkOrderReq) (resp *types.ChatWorkOrderResp, err error) {
	return logicutil.Proxy[types.ChatWorkOrderResp](l.ctx, req, l.svcCtx.ChatAdminCli.CreateChatWorkOrder)
}
