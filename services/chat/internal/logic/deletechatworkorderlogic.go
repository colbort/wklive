package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteChatWorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatWorkOrderLogic {
	return &DeleteChatWorkOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除工单
func (l *DeleteChatWorkOrderLogic) DeleteChatWorkOrder(in *chat.DeleteChatWorkOrderReq) (*chat.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminCommonResp{}, nil
}
