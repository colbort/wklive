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

type DeleteChatWorkOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatWorkOrderLogic {
	return &DeleteChatWorkOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteChatWorkOrderLogic) DeleteChatWorkOrder(req *types.DeleteChatWorkOrderReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.ChatAdminCli.DeleteChatWorkOrder)
}
