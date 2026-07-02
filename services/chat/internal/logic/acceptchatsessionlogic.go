package logic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
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
func (l *AcceptChatSessionLogic) AcceptChatSession(in *chat.AcceptChatSessionReq) (*chat.AcceptChatSessionResp, error) {
	agent, err := l.svcCtx.ChatAgentModel.FindOne(l.ctx, in.AgentId)
	if err == models.ErrNotFound {
		return &chat.AcceptChatSessionResp{Base: helper.ErrResp(404, "chat agent not found")}, nil
	}
	if err != nil {
		return &chat.AcceptChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if agent.Status != int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_ONLINE) {
		return &chat.AcceptChatSessionResp{Base: helper.ErrResp(400, "chat agent is not online")}, nil
	}
	session, err := ih.AcceptChatSession(l.ctx, l.svcCtx, ih.AssignSessionOptions{
		SessionNo:  in.SessionNo,
		Agent:      agent,
		AssignType: chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL,
		Reason:     firstNonEmpty(in.GetReason(), "accept"),
		IsGuest:    in.GetIsGuest(),
	})
	if err != nil {
		return &chat.AcceptChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	user, err := l.svcCtx.ChatUserModel.FindOne(l.ctx, agent.UserId)
	if err != nil {
		return &chat.AcceptChatSessionResp{Base: helper.ErrResp(404, "chat user not found")}, nil
	}
	payload := &chat.ChatAgentPayload{
		SessionNo:     in.SessionNo,
		AgentId:       agent.Id,
		AgentUserId:   user.Id,
		AgentName:     user.Nickname,
		AgentAvatar:   user.AvatarUrl,
		AgentStatus:   chat.ChatAgentStatus(agent.Status),
		AssignType:    chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL,
		SessionStatus: chat.ChatSessionStatus(session.Status),
		Remark:        firstNonEmpty(in.GetReason(), "accept"),
		ActionTime:    utils.NowMillis(),
	}
	err = ih.PublishMessageEvent(l.ctx, l.svcCtx.BusRedis, chat.ChatAppEventChannel, ih.PublishEventAgentAccepted, &chat.ChatWsResponse_Agent{Agent: payload})
	if err != nil {
		return &chat.AcceptChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AcceptChatSessionResp{Base: helper.OkResp(), Data: &chat.AcceptChatSessionUser{
		SessionNo: in.SessionNo,
		User:      ih.SessionMessageUser(session),
		Agent:     payload,
	}}, nil
}

func agentServiceMessage(ctx context.Context, svcCtx *svc.ServiceContext, agent *models.TChatAgent) string {
	name := ""
	if svcCtx != nil && agent != nil && agent.UserId > 0 {
		user, err := svcCtx.ChatUserModel.FindOne(ctx, agent.UserId)
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
