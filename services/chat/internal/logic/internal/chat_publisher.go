package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		CreatedAt: utils.NowMillis(),
		Data:      ToProtoMessage(msg),
		Session:   ToProtoSession(session, false),
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

func PublishTransientMessageEvent(ctx context.Context, svcCtx *svc.ServiceContext, merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, channel string) error {
	if msg == nil {
		return fmt.Errorf("message data is empty")
	}
	if session == nil && svcCtx != nil && svcCtx.BusRedis != nil {
		session, _ = GetTransientSession(ctx, svcCtx.BusRedis, merchantId, msg.GetSessionNo())
	}
	eventType := msg.GetEventType()
	if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN {
		eventType = chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE
		msg.EventType = eventType
	}
	event := &chat.ChatMessageEvent{
		Type:      eventType,
		CreatedAt: utils.NowMillis(),
		Data:      msg,
		Session:   session,
	}
	if eventType != chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE {
		event.SessionEvent = transientSessionEventFromMessage(merchantId, msg, session, event.CreatedAt)
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
	msg := &chat.ChatMessage{
		MessageNo:   fmt.Sprintf("GM%d", now),
		SessionNo:   sessionNo,
		EventType:   eventType,
		MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT,
		Content:     strings.TrimSpace(message),
		Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		AgentId:     strconv.FormatInt(agentId, 10),
		CreateTime:  now,
		UpdateTime:  now,
		Sender: &chat.ChatMessageUser{
			Id:       userId,
			Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
			Nickname: "系统",
		},
	}
	event := &chat.ChatMessageEvent{
		Type:      eventType,
		CreatedAt: now,
		Data:      msg,
		Session:   session,
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  sessionNo,
			MerchantId: merchantId,
			UserId:     userId,
			AgentId:    agentId,
			Message:    strings.TrimSpace(message),
			Session:    session,
			CreatedAt:  now,
		},
	}
	if session != nil {
		event.SessionEvent.Status = session.GetStatus()
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
	messageNo, err := svcCtx.GenerateNo(ctx, "GM")
	if err != nil {
		logx.Errorf("generate message no error: %v", err)
		return
	}
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE,
		CreatedAt: utils.NowMillis(),
		Data:      NewEventSystemMessage(messageNo, session, queue.GetMessage()),
		Session:   ToProtoSession(session, false),
		Queue:     queue,
	}
	publishChatEvent(ctx, svcCtx, event, "publish chat queue event failed", channel)
}

func PublishSessionEvent(ctx context.Context, svcCtx *svc.ServiceContext, eventType chat.ChatEventType, isGuest bool, session *models.TChatSession, assignType chat.ChatAssignType, reason, message string, channel string) {
	if svcCtx.BusRedis == nil || session == nil {
		return
	}
	queue, err := ToProtoQueueInfo(ctx, svcCtx, session)
	if err != nil {
		logx.WithContext(ctx).Errorf("build chat session event queue failed: %v", err)
	}
	sessionEvent := &chat.ChatSessionEvent{
		SessionNo:  session.SessionNo,
		MerchantId: session.MerchantId,
		UserId:     session.UserId,
		AgentId:    session.AgentId,
		OperatorId: session.UserId,
		Status:     chat.ChatSessionStatus(session.Status),
		AssignType: assignType,
		Reason:     strings.TrimSpace(reason),
		Message:    strings.TrimSpace(message),
		Session:    ToProtoSession(session, isGuest),
		Queue:      queue,
		CreatedAt:  utils.NowMillis(),
	}
	messageNo, err := svcCtx.GenerateNo(ctx, "GM")
	if err != nil {
		logx.Errorf("generate message no error: %v", err)
		return
	}
	event := &chat.ChatMessageEvent{
		Type:         eventType,
		CreatedAt:    sessionEvent.CreatedAt,
		Data:         NewEventSystemMessage(messageNo, session, sessionEvent.GetMessage()),
		Session:      sessionEvent.Session,
		SessionEvent: sessionEvent,
		Queue:        queue,
	}
	publishChatEvent(ctx, svcCtx, event, "publish chat session event failed", channel)
}

func PublishAgentStatusEvent(ctx context.Context, svcCtx *svc.ServiceContext, agent *models.TChatAgent) {
	if svcCtx.BusRedis == nil || agent == nil {
		return
	}
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_JOIN,
		CreatedAt: utils.NowMillis(),
		Agent:     ToProtoAgent(agent),
	}
	publishChatEvent(ctx, svcCtx, event, "publish chat agent status event failed", chat.ChatAdminEventChannel)
}

func PublishEvaluationEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, satisfaction *models.TChatSatisfaction, channel string) {
	if svcCtx.BusRedis == nil || session == nil || satisfaction == nil {
		return
	}
	messageNo, err := svcCtx.GenerateNo(ctx, "GM")
	if err != nil {
		logx.Errorf("generate message no error: %v", err)
		return
	}
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT,
		CreatedAt: utils.NowMillis(),
		Data: &chat.ChatMessage{
			MessageNo:   messageNo,
			SessionNo:   session.SessionNo,
			EventType:   chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT,
			MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_EVALUATION,
			Content:     "用户已提交评价",
			Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
			AgentId:     strconv.FormatInt(session.AgentId, 10),
			Extra:       SatisfactionPayloadJSON(satisfaction),
			CreateTime:  utils.NowMillis(),
			UpdateTime:  utils.NowMillis(),
			Sender: &chat.ChatMessageUser{
				Id:       session.UserId,
				Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
				Nickname: session.Title,
			},
		},
		Session: ToProtoSession(session, false),
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  session.SessionNo,
			MerchantId: session.MerchantId,
			UserId:     session.UserId,
			AgentId:    session.AgentId,
			Status:     chat.ChatSessionStatus(session.Status),
			Message:    "用户已提交评价",
			Session:    ToProtoSession(session, false),
			CreatedAt:  utils.NowMillis(),
		},
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
		EventType:   chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM,
		MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT,
		Content:     strings.TrimSpace(content),
		Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		AgentId:     strconv.FormatInt(session.AgentId, 10),
		CreateTime:  now,
		UpdateTime:  now,
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

func transientSessionEventFromMessage(merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatSessionEvent {
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
	event := &chat.ChatSessionEvent{
		SessionNo:  msg.GetSessionNo(),
		MerchantId: merchantId,
		UserId:     userId,
		AgentId:    int64FromProtoString(msg.GetAgentId()),
		Message:    strings.TrimSpace(msg.GetContent()),
		Session:    session,
		CreatedAt:  createdAt,
	}
	if session != nil {
		event.Status = session.GetStatus()
		if event.AgentId <= 0 {
			event.AgentId = session.GetAgentId()
		}
	}
	return event
}

func int64FromProtoString(value string) int64 {
	id, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return id
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
