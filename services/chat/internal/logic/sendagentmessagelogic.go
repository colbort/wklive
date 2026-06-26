package logic

import (
	"context"
	"fmt"

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
	session, base, err := getSession(l.ctx, l.svcCtx, in.MerchantId, in.SessionNo)
	if err != nil {
		return &chat.AdminChatMessageResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AdminChatMessageResp{Base: base}, nil
	}
	agent, err := l.svcCtx.ChatAgentModel.FindOneByMerchantIdChatUserId(l.ctx, in.MerchantId, in.Sender.Id)
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
	now := nowMillis()
	msg, err := sendMessage(l.ctx, l.svcCtx, session, &models.ChatMessage{
		MessageNo:  nextNo("CM"),
		SessionNo:  session.SessionNo,
		MerchantId: session.MerchantId,
		Sender: &models.ChatMessageUser{
			Id:        in.Sender.Id,
			Type:      int64(in.Sender.Type),
			Nickname:  in.Sender.Nickname,
			AvatarUrl: in.Sender.AvatarUrl,
		},
		Receiver: &models.ChatMessageUser{
			Id:        session.UserId,
			Type:      int64(chat.ChatSenderType_CHAT_SENDER_TYPE_USER),
			Nickname:  fmt.Sprintf("User%d", session.Id),
			AvatarUrl: "",
		},
		MessageType: int64(in.MessageType),
		Content:     in.Content,
		Url:         in.Url,
		FileName:    in.FileName,
		MimeType:    in.MimeType,
		FileSize:    in.FileSize,
		Duration:    in.Duration,
		Status:      int64(chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT),
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		return &chat.AdminChatMessageResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatMessageResp{Base: okBase(), Data: toProtoMessage(msg)}, nil
}
