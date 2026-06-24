package logic

import (
	"context"
	"fmt"
	"strings"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

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
	agent, err := l.svcCtx.ChatAgentModel.FindOne(l.ctx, in.GetAgentId())
	if err == models.ErrNotFound {
		return &chat.AdminChatSessionResp{Base: notFoundBase("chat agent not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatSessionResp{Base: errorBase(err)}, nil
	}
	if agent.Status != int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_ONLINE) {
		return &chat.AdminChatSessionResp{Base: badBase("chat agent is not online")}, nil
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
	publishSessionEvent(l.ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_ACCEPTED, session, operatorID, chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL, in.GetReason(), agentServiceMessage(l.ctx, l.svcCtx, agent))
	publishQueueEvent(l.ctx, l.svcCtx, session)
	return &chat.AdminChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}

func agentServiceMessage(ctx context.Context, svcCtx *svc.ServiceContext, agent *models.TChatAgent) string {
	name := ""
	if svcCtx != nil && agent != nil && agent.ChatUserId > 0 {
		user, err := svcCtx.ChatUserModel.FindOne(ctx, agent.ChatUserId)
		if err == nil && user != nil && user.MerchantId == agent.MerchantId {
			name = strings.TrimSpace(user.Nickname)
		}
	}
	if name == "" && agent != nil {
		name = strings.TrimSpace(agent.AgentNo)
	}
	if name == "" {
		return "客服正在为你服务"
	}
	return fmt.Sprintf("%s 客服正在为你服务", name)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
