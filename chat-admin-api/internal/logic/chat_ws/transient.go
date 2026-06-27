package chat_ws

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wklive/proto/chat"

	"google.golang.org/protobuf/types/known/structpb"
)

func newTransientAgentMessage(merchantId int64, sessionNo string, userId int64, agentId int64, senderNickname string, data sendAgentMessagePayload) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	senderNickname = firstNonEmpty(data.SenderNickname, senderNickname)
	return &chat.ChatMessage{
		MessageNo:   nextTransientNo("GM"),
		SessionNo:   sessionNo,
		MerchantId:  merchantId,
		MessageType: chat.ChatMessageType(data.MessageType),
		Sender: &chat.ChatMessageUser{
			Id:        agentId,
			Type:      chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT,
			Nickname:  senderNickname,
			AvatarUrl: strings.TrimSpace(data.SenderAvatarUrl),
		},
		Receiver: &chat.ChatMessageUser{
			Id:   userId,
			Type: chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
		},
		Content:     strings.TrimSpace(data.Content),
		Url:         strings.TrimSpace(data.Url),
		FileName:    strings.TrimSpace(data.FileName),
		MimeType:    strings.TrimSpace(data.MimeType),
		FileSize:    data.FileSize,
		Width:       int32(data.Width),
		Height:      int32(data.Height),
		Duration:    data.Duration,
		Payload:     structFromJSONString(data.Extra),
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
		Sender: &chat.ChatMessageUser{
			Id:       agentId,
			Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
			Nickname: "系统",
		},
		Receiver: &chat.ChatMessageUser{
			Id:   userId,
			Type: chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
		},
		MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT,
		Content:     strings.TrimSpace(content),
		Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		CreateTimes: now,
		UpdateTimes: now,
	}
}

func newTransientSystemMessageWithType(merchantId int64, sessionNo string, userId int64, agentId int64, _ chat.ChatEventType, messageType chat.ChatMessageType, content string, extra string) *chat.ChatMessage {
	msg := newTransientSystemMessage(merchantId, sessionNo, userId, agentId, content)
	msg.MessageType = messageType
	msg.Payload = structFromJSONString(extra)
	return msg
}

func structFromJSONString(value string) *structpb.Struct {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil
	}
	st, err := structpb.NewStruct(payload)
	if err != nil {
		return nil
	}
	return st
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
