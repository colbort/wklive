package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatWorkOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatWorkOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatWorkOrdersLogic {
	return &PageChatWorkOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询工单
func (l *PageChatWorkOrdersLogic) PageChatWorkOrders(in *chat.PageChatWorkOrdersReq) (*chat.PageChatWorkOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &chat.PageChatWorkOrdersResp{}, nil
}
