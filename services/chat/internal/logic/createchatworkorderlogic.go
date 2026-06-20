package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatWorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatWorkOrderLogic {
	return &CreateChatWorkOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建工单/离线留言
func (l *CreateChatWorkOrderLogic) CreateChatWorkOrder(in *chat.CreateChatWorkOrderReq) (*chat.AdminChatWorkOrderResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatWorkOrderResp{}, nil
}
