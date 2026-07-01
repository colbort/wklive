package helper

import (
	"context"
	"fmt"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"
)

func MessageNextCursor(list []*models.ChatMessage) int64 {
	if len(list) == 0 {
		return 0
	}
	return list[len(list)-1].CreateTimes
}

func MessageSenderType(msg *models.ChatMessage) chat.ChatSenderType {
	if msg == nil || msg.Sender == nil {
		return chat.ChatSenderType_CHAT_SENDER_TYPE_UNKNOWN
	}
	return chat.ChatSenderType(msg.Sender.Type)
}

func MessageAgentID(msg *models.ChatMessage) int64 {
	if msg == nil {
		return 0
	}
	if msg.Sender != nil && chat.ChatSenderType(msg.Sender.Type) == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		return msg.Sender.Id
	}
	if msg.Receiver != nil && chat.ChatSenderType(msg.Receiver.Type) == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		return msg.Receiver.Id
	}
	return 0
}

func MessageReceiver(session *models.TChatSession, senderType chat.ChatSenderType) *models.ChatMessageUser {
	if session == nil {
		return nil
	}
	switch senderType {
	case chat.ChatSenderType_CHAT_SENDER_TYPE_USER:
		if session.AgentId <= 0 {
			return nil
		}
		return &models.ChatMessageUser{
			Id:   session.AgentId,
			Type: int64(chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT),
		}
	case chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT, chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM:
		if session.UserId <= 0 {
			return nil
		}
		return &models.ChatMessageUser{
			Id:   session.UserId,
			Type: int64(chat.ChatSenderType_CHAT_SENDER_TYPE_USER),
		}
	default:
		return nil
	}
}

func SendMessage(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, msg *models.ChatMessage, channel string) (*models.ChatMessage, error) {
	model := svcCtx.ChatMessageFactory.New(session.MerchantId)
	if model == nil {
		return nil, fmt.Errorf("invalid merchant_id: %d", session.MerchantId)
	}
	if err := model.Insert(ctx, msg); err != nil {
		return nil, err
	}

	now := msg.CreateTimes
	session.LastMessageNo = msg.MessageNo
	session.LastMessage = TrimSummary(msg.Content, msg.FileName, msg.Url)
	senderType := MessageSenderType(msg)
	session.LastSenderType = int64(senderType)
	session.LastMessageTime = now
	session.UpdateTimes = now
	switch senderType {
	case chat.ChatSenderType_CHAT_SENDER_TYPE_USER:
		session.AgentUnreadCount++
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
			session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT)
		}
	case chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT:
		session.UserUnreadCount++
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
			session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER)
		}
	case chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM:
		session.UserUnreadCount++
	}
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return nil, err
	}
	_ = PublishMessageEvent(ctx, svcCtx, PublishMessageEventReq{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		Channel:   channel,
		Message:   msg,
		Session:   session,
	})
	return msg, nil
}
