package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AcceptChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAcceptChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptChatSessionLogic {
	return &AcceptChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 接待会话
func (l *AcceptChatSessionLogic) AcceptChatSession(in *chat.AcceptChatSessionReq) (*chat.AdminChatSessionResp, error) {
	if in.GetAgentId() <= 0 {
		return &chat.AdminChatSessionResp{Base: badBase("agent_id is required")}, nil
	}
	operatorID := in.GetOperatorId()
	if operatorID <= 0 {
		operatorID = in.GetAgentId()
	}
	session, base, err := assignSession(l.ctx, l.svcCtx, &chat.AssignChatSessionReq{
		SessionNo:  in.GetSessionNo(),
		ToAgentId:  in.GetAgentId(),
		AssignType: chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL,
		OperatorId: operatorID,
		Reason:     firstNonEmpty(in.GetReason(), "accept"),
	})
	if err != nil {
		return &chat.AdminChatSessionResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AdminChatSessionResp{Base: base}, nil
	}
	publishSessionEvent(l.ctx, l.svcCtx, chat.ChatMessageEventTypeSessionAccepted, session, operatorID, chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL, in.GetReason(), "客服已接入")
	publishQueueEvent(l.ctx, l.svcCtx, session)
	return &chat.AdminChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
