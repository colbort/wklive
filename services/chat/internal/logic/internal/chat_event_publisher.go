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

func PublishMessageEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, msg *models.ChatMessage) {
	if svcCtx.BusRedis == nil || msg == nil {
		return
	}

	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		CreatedAt: utils.NowMillis(),
		Data:      ToProtoMessage(msg),
		Session:   ToProtoSession(session),
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat message event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat message event failed: %v", err)
	}
}

func PublishChatEvent(ctx context.Context, svcCtx *svc.ServiceContext, event *chat.ChatMessageEvent) error {
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
	_, err = svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload))
	return err
}

func PublishQueueEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) {
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
		Session:   ToProtoSession(session),
		Queue:     queue,
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat queue event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat queue event failed: %v", err)
	}
}

func PublishSessionEvent(ctx context.Context, svcCtx *svc.ServiceContext, eventType chat.ChatEventType, session *models.TChatSession, assignType chat.ChatAssignType, reason, message string) {
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
		Session:    ToProtoSession(session),
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
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat session event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat session event failed: %v", err)
	}
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
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat agent status event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat agent status event failed: %v", err)
	}
}

func PublishEvaluationEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, satisfaction *models.TChatSatisfaction) {
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
		Session: ToProtoSession(session),
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  session.SessionNo,
			MerchantId: session.MerchantId,
			UserId:     session.UserId,
			AgentId:    session.AgentId,
			Status:     chat.ChatSessionStatus(session.Status),
			Message:    "用户已提交评价",
			Session:    ToProtoSession(session),
			CreatedAt:  utils.NowMillis(),
		},
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat evaluation event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat evaluation event failed: %v", err)
	}
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
