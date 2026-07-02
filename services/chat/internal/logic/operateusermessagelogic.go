package logic

import (
	"context"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperateUserMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperateUserMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperateUserMessageLogic {
	return &OperateUserMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户侧消息删除/撤回
func (l *OperateUserMessageLogic) OperateUserMessage(in *chat.OperateUserMessageReq) (*chat.AppCommonResp, error) {
	return &chat.AppCommonResp{
		Base: ih.OperateMessage(l.ctx, l.svcCtx, in.GetMessageOperate(), chat.ChatSenderType_CHAT_SENDER_TYPE_USER, in.GetIsGuest()),
	}, nil
}
