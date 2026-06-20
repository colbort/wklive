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

type PageChatWorkOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageChatWorkOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatWorkOrdersLogic {
	return &PageChatWorkOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageChatWorkOrdersLogic) PageChatWorkOrders(req *types.PageChatWorkOrdersReq) (resp *types.PageChatWorkOrdersResp, err error) {
	return logicutil.Proxy[types.PageChatWorkOrdersResp](l.ctx, req, l.svcCtx.ChatAdminCli.PageChatWorkOrders)
}
