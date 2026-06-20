package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatWorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatWorkOrderLogic {
	return &UpdateChatWorkOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新工单
func (l *UpdateChatWorkOrderLogic) UpdateChatWorkOrder(in *chat.UpdateChatWorkOrderReq) (*chat.AdminChatWorkOrderResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatWorkOrderResp{}, nil
}
