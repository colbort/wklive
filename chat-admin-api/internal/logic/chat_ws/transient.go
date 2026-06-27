package chat_ws

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wklive/proto/chat"
)

func newTransientAgentMessage(_ int64, sessionNo string, userId int64, agentId int64, senderNickname string, data sendAgentMessagePayload) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	senderNickname = firstNonEmpty(data.SenderNickname, senderNickname)
	return &chat.ChatMessage{
		MessageNo:   nextTransientNo("GM"),
		SessionNo:   sessionNo,
		EventType:   chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		AgentId:     strconv.FormatInt(agentId, 10),
		MessageType: chat.ChatMessageType(data.MessageType),
		Sender: &chat.ChatMessageUser{
			Id:        userId,
			Type:      chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT,
			Nickname:  senderNickname,
			AvatarUrl: strings.TrimSpace(data.SenderAvatarUrl),
		},
		Content:    strings.TrimSpace(data.Content),
		Url:        strings.TrimSpace(data.Url),
		FileName:   strings.TrimSpace(data.FileName),
		MimeType:   strings.TrimSpace(data.MimeType),
		FileSize:   data.FileSize,
		Width:      int32(data.Width),
		Height:     int32(data.Height),
		Duration:   int32(data.Duration),
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

func transientExtra(payload map[string]interface{}) string {
	bs, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(bs)
}

func nextTransientNo(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s%d%s", prefix, time.Now().UnixMilli(), hex.EncodeToString(b))
}
