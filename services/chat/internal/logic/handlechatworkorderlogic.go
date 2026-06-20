package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleChatWorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleChatWorkOrderLogic {
	return &HandleChatWorkOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 处理工单
func (l *HandleChatWorkOrderLogic) HandleChatWorkOrder(in *chat.HandleChatWorkOrderReq) (*chat.AdminChatWorkOrderResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatWorkOrderResp{}, nil
}
