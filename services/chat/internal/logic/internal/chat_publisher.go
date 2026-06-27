package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
)

func publishChatEventToChannels(ctx context.Context, svcCtx *svc.ServiceContext, event *chat.ChatMessageEvent, channel string) error {
	if svcCtx.BusRedis == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	if event == nil {
		return fmt.Errorf("chat event is empty")
	}
	if event.GetCreatedAt() == 0 {
		event.CreatedAt = utils.NowMillis()
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, channel, string(payload)); err != nil {
		return err
	}
	return nil
}

func publishChatEvent(ctx context.Context, svcCtx *svc.ServiceContext, event *chat.ChatMessageEvent, logMessage string, channel string) {
	if svcCtx.BusRedis == nil || event == nil {
		return
	}
	if err := publishChatEventToChannels(ctx, svcCtx, event, channel); err != nil {
		logx.WithContext(ctx).Errorf("%s: %v", logMessage, err)
	}
}

func PublishMessageEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, msg *models.ChatMessage, channel string) {
	if svcCtx.BusRedis == nil || msg == nil {
		return
	}

	event := &chat.ChatMessageEvent{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		CreatedAt: utils.NowMillis(),
		Payload:   &chat.ChatMessageEvent_Message{Message: ToProtoMessage(msg)},
	}
	publishChatEvent(ctx, svcCtx, event, "publish chat message event failed", channel)
}

func PublishChatEvent(ctx context.Context, svcCtx *svc.ServiceContext, event *chat.ChatMessageEvent, channel string) error {
	if svcCtx.BusRedis == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	if event == nil {
		return fmt.Errorf("chat event is empty")
	}
	return publishChatEventToChannels(ctx, svcCtx, event, channel)
}

func PublishTransientMessageEvent(ctx context.Context, svcCtx *svc.ServiceContext, merchantId int64, eventType chat.ChatEventType, msg *chat.ChatMessage, session *chat.ChatSession, channel string) error {
	if msg == nil {
		return fmt.Errorf("message data is empty")
	}
	if session == nil && svcCtx != nil && svcCtx.BusRedis != nil {
		session, _ = GetTransientSession(ctx, svcCtx.BusRedis, merchantId, msg.GetSessionNo())
	}
	if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN {
		eventType = chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE
	}
	event := &chat.ChatMessageEvent{EventType: eventType, CreatedAt: utils.NowMillis()}
	if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE || eventType == chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM {
		event.Payload = &chat.ChatMessageEvent_Message{Message: msg}
	} else {
		event.Payload = &chat.ChatMessageEvent_Session{Session: transientSessionFromMessage(merchantId, eventType, msg, session, event.CreatedAt)}
	}
	return PublishChatEvent(ctx, svcCtx, event, channel)
}

func PublishTransientSessionEvent(ctx context.Context, svcCtx *svc.ServiceContext, eventType chat.ChatEventType, merchantId int64, session *chat.ChatSession, sessionNo string, userId, agentId int64, message, channel string) error {
	if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN {
		return nil
	}
	if session != nil {
		sessionNo = firstNonEmptyString(sessionNo, session.GetSessionNo())
		if merchantId <= 0 {
			merchantId = session.GetMerchantId()
		}
		if userId <= 0 {
			userId = session.GetUserId()
		}
		if agentId <= 0 {
			agentId = session.GetAgentId()
		}
	}
	if strings.TrimSpace(sessionNo) == "" {
		return fmt.Errorf("session_no is required")
	}
	now := utils.NowMillis()
	session = sessionWithEventMeta(session, eventType, strings.TrimSpace(message), "", chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN, now)
	if session == nil {
		session = sessionWithEventMeta(&chat.ChatSession{
			SessionNo:  sessionNo,
			MerchantId: merchantId,
			UserId:     userId,
			AgentId:    agentId,
		}, eventType, strings.TrimSpace(message), "", chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN, now)
	}
	event := &chat.ChatMessageEvent{
		EventType: eventType,
		CreatedAt: now,
		Payload:   &chat.ChatMessageEvent_Session{Session: session},
	}
	return PublishChatEvent(ctx, svcCtx, event, channel)
}

func PublishQueueEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, channel string) {
	if svcCtx.BusRedis == nil || session == nil {
		return
	}
	queue, err := ToProtoQueueInfo(ctx, svcCtx, session)
	if err != nil {
		logx.WithContext(ctx).Errorf("build chat queue event failed: %v", err)
		return
	}
	event := &chat.ChatMessageEvent{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE,
		CreatedAt: utils.NowMillis(),
		Payload:   &chat.ChatMessageEvent_Queue{Queue: queue},
	}
	publishChatEvent(ctx, svcCtx, event, "publish chat queue event failed", channel)
}

func PublishSessionEvent(ctx context.Context, svcCtx *svc.ServiceContext, eventType chat.ChatEventType, isGuest bool, session *models.TChatSession, assignType chat.ChatAssignType, reason, message string, channel string) {
	if svcCtx.BusRedis == nil || session == nil {
		return
	}
	now := utils.NowMillis()
	event := &chat.ChatMessageEvent{
		EventType: eventType,
		CreatedAt: now,
	}
	if isQueueEvent(eventType) {
		queue, err := ToProtoQueueInfo(ctx, svcCtx, session)
		if err != nil {
			logx.WithContext(ctx).Errorf("build chat session event queue failed: %v", err)
			return
		}
		event.Payload = &chat.ChatMessageEvent_Queue{Queue: queue}
	} else {
		event.Payload = &chat.ChatMessageEvent_Session{Session: sessionWithEventMeta(ToProtoSession(session, isGuest), eventType, strings.TrimSpace(message), strings.TrimSpace(reason), assignType, now)}
	}
	publishChatEvent(ctx, svcCtx, event, "publish chat session event failed", channel)
}

func PublishAgentStatusEvent(ctx context.Context, svcCtx *svc.ServiceContext, agent *models.TChatAgent) {
	if svcCtx.BusRedis == nil || agent == nil {
		return
	}
	event := &chat.ChatMessageEvent{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_JOIN,
		CreatedAt: utils.NowMillis(),
		Payload:   &chat.ChatMessageEvent_Agent{Agent: ToProtoAgent(agent)},
	}
	publishChatEvent(ctx, svcCtx, event, "publish chat agent status event failed", chat.ChatAdminEventChannel)
}

func PublishEvaluationEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, satisfaction *models.TChatSatisfaction, channel string) {
	if svcCtx.BusRedis == nil || session == nil || satisfaction == nil {
		return
	}
	event := &chat.ChatMessageEvent{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT,
		CreatedAt: utils.NowMillis(),
		Payload:   &chat.ChatMessageEvent_Satisfaction{Satisfaction: ToProtoSatisfaction(satisfaction)},
	}
	publishChatEvent(ctx, svcCtx, event, "publish chat evaluation event failed", channel)
}

func NewEventSystemMessage(messageNo string, session *models.TChatSession, content string) *chat.ChatMessage {
	if session == nil {
		return nil
	}
	now := utils.NowMillis()
	return &chat.ChatMessage{
		MessageNo:   messageNo,
		SessionNo:   session.SessionNo,
		MerchantId:  session.MerchantId,
		MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT,
		Content:     strings.TrimSpace(content),
		Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		CreateTimes: now,
		UpdateTimes: now,
		Sender: &chat.ChatMessageUser{
			Id:       session.AgentId,
			Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
			Nickname: "系统",
		},
	}
}

func SatisfactionPayloadJSON(data *models.TChatSatisfaction) string {
	if data == nil {
		return ""
	}
	payload := map[string]interface{}{
		"score":   data.Score,
		"content": data.Content,
		"tags":    data.Tags,
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(bs)
}

func transientSessionFromMessage(merchantId int64, eventType chat.ChatEventType, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatSession {
	if msg == nil {
		return nil
	}
	if session != nil && merchantId <= 0 {
		merchantId = session.GetMerchantId()
	}
	userId := msg.GetSender().GetId()
	if session != nil && userId <= 0 {
		userId = session.GetUserId()
	}
	if session == nil {
		session = &chat.ChatSession{
			SessionNo:  msg.GetSessionNo(),
			MerchantId: merchantId,
			UserId:     userId,
			AgentId:    transientMessageAgentID(msg, session),
		}
	}
	if session.GetSessionNo() == "" {
		session.SessionNo = msg.GetSessionNo()
	}
	if session.GetMerchantId() <= 0 {
		session.MerchantId = merchantId
	}
	if session.GetUserId() <= 0 {
		session.UserId = userId
	}
	if session.GetAgentId() <= 0 {
		session.AgentId = transientMessageAgentID(msg, session)
	}
	return sessionWithEventMeta(session, eventType, strings.TrimSpace(msg.GetContent()), "", chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN, createdAt)
}

func sessionWithEventMeta(session *chat.ChatSession, eventType chat.ChatEventType, message, reason string, assignType chat.ChatAssignType, createdAt int64) *chat.ChatSession {
	if session == nil {
		return nil
	}
	payload := map[string]interface{}{
		"eventType":       int64(eventType),
		"eventMessage":    strings.TrimSpace(message),
		"eventReason":     strings.TrimSpace(reason),
		"eventAssignType": int64(assignType),
		"eventCreatedAt":  createdAt,
	}
	if session.GetExtJson() != nil {
		for key, value := range session.GetExtJson().AsMap() {
			payload[key] = value
		}
	}
	session.ExtJson = MapToStruct(payload)
	return session
}

func isQueueEvent(eventType chat.ChatEventType) bool {
	return eventType == chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_JOIN ||
		eventType == chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE
}

func transientMessageAgentID(msg *chat.ChatMessage, session *chat.ChatSession) int64 {
	if session != nil && session.GetAgentId() > 0 {
		return session.GetAgentId()
	}
	if msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		return msg.GetSender().GetId()
	}
	return 0
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
