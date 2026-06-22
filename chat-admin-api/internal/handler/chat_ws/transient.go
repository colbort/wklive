package chat_ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"chat-admin-api/internal/svc"
	"wklive/proto/chat"

	"google.golang.org/protobuf/encoding/protojson"
)

const guestSessionPrefix = "GS"

func isGuestSession(sessionNo string) bool {
	return strings.HasPrefix(strings.TrimSpace(sessionNo), guestSessionPrefix)
}

func newTransientAgentMessage(merchantId int64, sessionNo string, userId int64, agentId int64, senderNickname string, data sendAgentMessagePayload) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	senderNickname = firstNonEmpty(data.SenderNickname, senderNickname)
	return &chat.ChatMessage{
		MessageNo:  nextTransientNo("GM"),
		SessionNo:  sessionNo,
		MerchantId: merchantId,
		UserId:     userId,
		AgentId:    agentId,
		SenderType: chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT,
		Sender: &chat.ChatMessageSender{
			Id:        agentId,
			Type:      chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT,
			Nickname:  senderNickname,
			AvatarUrl: strings.TrimSpace(data.SenderAvatarUrl),
		},
		MessageType: chat.ChatMessageType(data.MessageType),
		Content:     strings.TrimSpace(data.Content),
		MediaUrl:    strings.TrimSpace(data.MediaUrl),
		MediaName:   strings.TrimSpace(data.MediaName),
		MediaMime:   strings.TrimSpace(data.MediaMime),
		MediaSize:   data.MediaSize,
		Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		CreateTimes: now,
		UpdateTimes: now,
	}
}

func newTransientSystemMessage(merchantId int64, sessionNo string, userId int64, agentId int64, content string) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	return &chat.ChatMessage{
		MessageNo:  nextTransientNo("GM"),
		SessionNo:  sessionNo,
		MerchantId: merchantId,
		UserId:     userId,
		AgentId:    agentId,
		SenderType: chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
		Sender: &chat.ChatMessageSender{
			Id:       agentId,
			Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
			Nickname: "系统",
		},
		MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT,
		Content:     strings.TrimSpace(content),
		Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		CreateTimes: now,
		UpdateTimes: now,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}

func publishTransientMessage(ctx context.Context, svcCtx *svc.ServiceContext, msg *chat.ChatMessage) error {
	return publishTransientEvent(ctx, svcCtx, chat.ChatMessageEventTypeMessage, msg)
}

func publishTransientEvent(ctx context.Context, svcCtx *svc.ServiceContext, eventType string, msg *chat.ChatMessage) error {
	if svcCtx.BusRedis == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	event := &chat.ChatMessageEvent{
		Type:      eventType,
		Data:      msg,
		CreatedAt: time.Now().UnixMilli(),
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		return err
	}
	_, err = svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload))
	return err
}

func nextTransientNo(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s%d%s", prefix, time.Now().UnixMilli(), hex.EncodeToString(b))
}
