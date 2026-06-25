package logic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendAgentMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendAgentMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAgentMessageLogic {
	return &SendAgentMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送客服消息
func (l *SendAgentMessageLogic) SendAgentMessage(in *chat.SendAgentMessageReq) (*chat.AdminChatMessageResp, error) {
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatMessageResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatMessageResp{Base: errorBase(err)}, nil
	}
	session, base, err := getSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.AdminChatMessageResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AdminChatMessageResp{Base: base}, nil
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil || operatorID <= 0 {
		return &chat.AdminChatMessageResp{Base: badBase("operator_id is required")}, nil
	}
	agent, err := l.svcCtx.ChatAgentModel.FindOneByMerchantIdChatUserId(l.ctx, merchantID, operatorID)
	if err == models.ErrNotFound {
		return &chat.AdminChatMessageResp{Base: notFoundBase("chat agent not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatMessageResp{Base: errorBase(err)}, nil
	}
	if session.AgentId != 0 && session.AgentId != agent.Id {
		return &chat.AdminChatMessageResp{Base: badBase("agent does not own this session")}, nil
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return &chat.AdminChatMessageResp{Base: badBase("chat session is closed")}, nil
	}
	if session.AgentId == 0 {
		return &chat.AdminChatMessageResp{Base: badBase("chat session is not accepted")}, nil
	}
	msg := newMessage(session, chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT, agent.Id, "", "", in.GetMessageType(), in.GetContent(), in.GetUrl(), in.GetFileName(), in.GetMimeType(), in.GetFileSize(), nil)
	msg, err = sendMessage(l.ctx, l.svcCtx, session, msg)
	if err != nil {
		return &chat.AdminChatMessageResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatMessageResp{Base: okBase(), Data: toProtoMessage(msg)}, nil
}
