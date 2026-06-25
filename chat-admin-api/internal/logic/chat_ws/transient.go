package chat_ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/ws"
	"wklive/proto/chat"

	"google.golang.org/protobuf/encoding/protojson"
)

func newTransientAgentMessage(_ int64, sessionNo string, _ int64, agentId int64, senderNickname string, data sendAgentMessagePayload) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	senderNickname = firstNonEmpty(data.SenderNickname, senderNickname)
	return &chat.ChatMessage{
		MessageNo:   nextTransientNo("GM"),
		SessionNo:   sessionNo,
		EventType:   chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		AgentId:     strconv.FormatInt(agentId, 10),
		MessageType: chat.ChatMessageType(data.MessageType),
		Sender: &chat.ChatMessageUser{
			Id:        agentId,
			Type:      chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT,
			Nickname:  senderNickname,
			AvatarUrl: strings.TrimSpace(data.SenderAvatarUrl),
		},
		Content:    strings.TrimSpace(data.Content),
		Url:        strings.TrimSpace(data.MediaUrl),
		FileName:   strings.TrimSpace(data.MediaName),
		MimeType:   strings.TrimSpace(data.MediaMime),
		FileSize:   data.MediaSize,
		Extra:      strings.TrimSpace(data.Extra),
		Status:     chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		CreateTime: now,
		UpdateTime: now,
	}
}

func newTransientSystemMessage(_ int64, sessionNo string, _ int64, agentId int64, content string) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	return &chat.ChatMessage{
		MessageNo: nextTransientNo("GM"),
		SessionNo: sessionNo,
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM,
		AgentId:   strconv.FormatInt(agentId, 10),
		Sender: &chat.ChatMessageUser{
			Id:       agentId,
			Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
			Nickname: "系统",
		},
		MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT,
		Content:     strings.TrimSpace(content),
		Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		CreateTime:  now,
		UpdateTime:  now,
	}
}

func newTransientSystemMessageWithType(merchantId int64, sessionNo string, userId int64, agentId int64, eventType chat.ChatEventType, messageType chat.ChatMessageType, content string, extra string) *chat.ChatMessage {
	msg := newTransientSystemMessage(merchantId, sessionNo, userId, agentId, content)
	msg.EventType = eventType
	msg.MessageType = messageType
	msg.Extra = strings.TrimSpace(extra)
	return msg
}

func publishTransientMessage(ctx context.Context, svcCtx *svc.ServiceContext, msg *chat.ChatMessage) error {
	return publishTransientEvent(ctx, svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE, msg)
}

func publishTransientEvent(ctx context.Context, svcCtx *svc.ServiceContext, eventType chat.ChatEventType, msg *chat.ChatMessage) error {
	if svcCtx.BusRedis == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	event := &chat.ChatMessageEvent{
		Type:      eventType,
		Data:      msg,
		CreatedAt: time.Now().UnixMilli(),
	}
	if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED {
		merchantId, userId := transientMessageSessionIdentity(msg, svcCtx)
		event.SessionEvent = &chat.ChatSessionEvent{
			SessionNo:  msg.GetSessionNo(),
			MerchantId: merchantId,
			UserId:     userId,
			AgentId:    int64FromString(msg.GetAgentId()),
			Status:     chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING,
			Message:    msg.GetContent(),
			CreatedAt:  event.CreatedAt,
		}
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		return err
	}
	_, err = svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload))
	return err
}

func transientExtra(payload map[string]interface{}) string {
	bs, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(bs)
}

func transientMessageSessionIdentity(msg *chat.ChatMessage, svcCtx *svc.ServiceContext) (int64, int64) {
	if svcCtx == nil || svcCtx.ChatMessageHub == nil || msg == nil {
		return 0, 0
	}
	sessions := svcCtx.ChatMessageHub.ListTransientSessions(ws.TransientSessionFilter{})
	for _, session := range sessions {
		if session.GetSessionNo() == msg.GetSessionNo() {
			return session.GetMerchantId(), session.GetUserId()
		}
	}
	return 0, 0
}

func int64FromString(value string) int64 {
	id, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return id
}

func nextTransientNo(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s%d%s", prefix, time.Now().UnixMilli(), hex.EncodeToString(b))
}
