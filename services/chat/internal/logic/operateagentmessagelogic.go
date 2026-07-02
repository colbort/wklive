package logic

import (
	"context"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperateAgentMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperateAgentMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperateAgentMessageLogic {
	return &OperateAgentMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 客服侧消息删除/撤回
func (l *OperateAgentMessageLogic) OperateAgentMessage(in *chat.OperateAgentMessageReq) (*chat.AdminCommonResp, error) {
	return &chat.AdminCommonResp{
		Base: ih.OperateMessage(l.ctx, l.svcCtx, in.GetMessageOperate(), chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT, in.GetIsGuest()),
	}, nil
}
