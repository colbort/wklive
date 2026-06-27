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

func publishTransientMessage(ctx context.Context, svcCtx *svc.ServiceContext, merchantId int64, msg *chat.ChatMessage) error {
	return publishTransientEvent(ctx, svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE, merchantId, msg)
}

func publishTransientEvent(ctx context.Context, svcCtx *svc.ServiceContext, eventType chat.ChatEventType, merchantId int64, msg *chat.ChatMessage) error {
	if svcCtx == nil || svcCtx.ChatAdminCli == nil {
		return fmt.Errorf("chat admin client is not configured")
	}
	event := &chat.ChatMessageEvent{
		Type:      eventType,
		Data:      msg,
		CreatedAt: time.Now().UnixMilli(),
	}
	if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED {
		userId := transientMessageSessionUserId(ctx, msg, merchantId, svcCtx)
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
	resp, err := svcCtx.ChatAdminCli.AdminPublishChatEvent(ctx, &chat.AdminPublishChatEventReq{Event: event})
	if err != nil {
		return err
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("%s", resp.GetBase().GetMsg())
	}
	return nil
}

func transientExtra(payload map[string]interface{}) string {
	bs, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(bs)
}

func transientMessageSessionUserId(ctx context.Context, msg *chat.ChatMessage, merchantId int64, svcCtx *svc.ServiceContext) int64 {
	if svcCtx == nil || svcCtx.ChatAdminCli == nil || msg == nil || merchantId <= 0 {
		return 0
	}
	resp, err := svcCtx.ChatAdminCli.AdminGetTransientChatSession(ctx, &chat.AdminGetTransientChatSessionReq{
		MerchantId: merchantId,
		SessionNo:  msg.GetSessionNo(),
	})
	if err != nil || resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		return 0
	}
	return resp.GetData().GetUserId()
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
