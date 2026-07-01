package helper

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"google.golang.org/protobuf/encoding/protojson"
)

type PublishMessageEventReq struct {
	EventType    chat.ChatEventType
	Channel      string
	MerchantId   int64
	SessionNo    string
	UserId       int64
	AgentId      int64
	IsGuest      bool
	AssignType   chat.ChatAssignType
	Reason       string
	EventMessage string
	Message      *models.ChatMessage
	Session      *models.TChatSession
	Agent        *models.TChatAgent
	Satisfaction *models.TChatSatisfaction
}

func PublishMessageEvent(ctx context.Context, svcCtx *svc.ServiceContext, req PublishMessageEventReq) error {
	if req.EventType == chat.ChatEventType_CHAT_EVENT_TYPE_UNSPECIFIED {
		return fmt.Errorf("event_type is required")
	}
	createdAt := utils.NowMillis()
	event := &chat.ChatMessageEvent{
		Code:      200,
		Msg:       "",
		EventType: req.EventType,
		CreatedAt: createdAt,
	}
	switch req.EventType {
	case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE:
		event.Payload = &chat.ChatMessageEvent_Message{Message: &chat.ChatMessage{}}
	case chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM_NOTICE:
		event.Payload = &chat.ChatMessageEvent_SystemNotice{SystemNotice: &chat.ChatSystemNoticePayload{
			SessionNo:  req.SessionNo,
			Title:      "",
			Content:    strings.TrimSpace(req.EventMessage),
			Level:      "info",
			ShowInChat: false,
		}}
	case chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE:
		event.Payload = &chat.ChatMessageEvent_Queue{Queue: &chat.ChatQueuePayload{
			SessionNo:            req.SessionNo,
			UserId:               req.UserId,
			Nickname:             "",
			QueueAction:          chat.ChatQueueAction_CHAT_QUEUE_ACTION_JOIN,
			QueuePosition:        0,
			WaitingCount:         0,
			EstimatedWaitSeconds: 0,
			SessionStatus:        chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT,
			ActionTime:           utils.NowMillis(),
		}}
	case chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ACCEPTED,
		chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_LEAVE:
		event.Payload = &chat.ChatMessageEvent_Agent{Agent: &chat.ChatAgentPayload{
			SessionNo:     req.SessionNo,
			AgentId:       req.AgentId,
			AgentName:     "",
			AgentAvatar:   "",
			AgentStatus:   chat.ChatAgentStatus(req.Agent.Status),
			AssignType:    req.AssignType,
			SessionStatus: chat.ChatSessionStatus(req.Session.Status),
			Remark:        strings.TrimSpace(req.EventMessage),
			ActionTime:    createdAt,
		}}
	case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE:
		event.Payload = &chat.ChatMessageEvent_Evaluation{Evaluation: evaluationInviteFromMessage(req.MerchantId, req.Message, req.Session, createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT:
		if req.Satisfaction == nil {
			return fmt.Errorf("satisfaction data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Evaluation{Evaluation: evaluationSubmitPayload(req.Satisfaction)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING:
		event.Payload = &chat.ChatMessageEvent_Typing{Typing: typingFromMessage(req.MerchantId, req.Message, req.Session, createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELIVERED,
		chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_READ:
		event.Payload = &chat.ChatMessageEvent_Receipt{Receipt: receiptFromMessage(req.MerchantId, req.Message, req.Session, createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL,
		chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELETE:
		event.Payload = &chat.ChatMessageEvent_MessageOperate{MessageOperate: messageOperatePayload(req, createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT:
		event.Payload = &chat.ChatMessageEvent_Heartbeat{Heartbeat: &chat.ChatHeartbeatPayload{ServerTime: createdAt}}
	case chat.ChatEventType_CHAT_EVENT_TYPE_ERROR:
		event.Payload = &chat.ChatMessageEvent_Error{Error: chatEventErrorFromMessage(req.MerchantId, req.Message, req.Session, createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN,
		chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE,
		chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REQUEST,
		chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_ACCEPT,
		chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REJECT,
		chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE:
		event.Payload = &chat.ChatMessageEvent_Session{Session: ToProtoSession(req.Session, req.IsGuest)}
	default:
		return fmt.Errorf("unsupported chat event type: %s", req.EventType.String())
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, req.Channel, string(payload)); err != nil {
		return err
	}
	return nil
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

func chatEventErrorFromMessage(merchantId int64, msg *models.ChatMessage, session *models.TChatSession, createdAt int64) *chat.ChatErrorPayload {
	if session != nil && merchantId <= 0 {
		merchantId = session.MerchantId
	}
	_ = merchantId
	_ = createdAt
	return &chat.ChatErrorPayload{
		SessionNo:    msg.SessionNo,
		ErrorCode:    500,
		ErrorMessage: strings.TrimSpace(msg.Content),
		Retryable:    false,
	}
}

func evaluationInviteFromMessage(merchantId int64, msg *models.ChatMessage, session *models.TChatSession, createdAt int64) *chat.ChatEvaluationPayload {
	if msg == nil {
		return nil
	}
	if session != nil && merchantId <= 0 {
		merchantId = session.MerchantId
	}
	invite := &chat.ChatEvaluationPayload{
		SessionNo:   msg.SessionNo,
		Comment:     strings.TrimSpace(msg.Content),
		Submitted:   false,
		EvaluatedAt: createdAt,
	}
	_ = merchantId
	if session != nil {
		invite.UserId = session.UserId
		invite.AgentId = session.AgentId
	}
	if sender := msg.Sender; sender != nil && chat.ChatSenderType(sender.Type) == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		invite.AgentId = sender.Id
	}
	if receiver := msg.Receiver; receiver != nil && chat.ChatSenderType(receiver.Type) == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
		invite.UserId = receiver.Id
	}
	return invite
}

func evaluationSubmitPayload(data *models.TChatSatisfaction) *chat.ChatEvaluationPayload {
	if data == nil {
		return nil
	}
	tags := make([]string, 0)
	for _, tag := range strings.Split(data.Tags, ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	return &chat.ChatEvaluationPayload{
		SessionNo:    data.SessionNo,
		UserId:       data.UserId,
		AgentId:      data.AgentId,
		EvaluationId: data.Id,
		Rating:       int32(data.Score),
		Tags:         tags,
		Comment:      data.Content,
		Submitted:    true,
		EvaluatedAt:  data.UpdateTimes,
	}
}

func typingFromMessage(merchantId int64, msg *models.ChatMessage, session *models.TChatSession, createdAt int64) *chat.ChatTypingPayload {
	if msg == nil {
		return nil
	}
	if session != nil && merchantId <= 0 {
		merchantId = session.MerchantId
	}
	userId := int64(0)
	agentId := int64(0)
	if session != nil {
		userId = session.UserId
		agentId = session.AgentId
	}
	if msg.Sender != nil {
		if chat.ChatSenderType(msg.Sender.Type) == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
			userId = msg.Sender.Id
		}
		if chat.ChatSenderType(msg.Sender.Type) == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
			agentId = msg.Sender.Id
		}
	}
	typing := &chat.ChatTypingPayload{
		SessionNo:  msg.SessionNo,
		SenderId:   userId,
		Text:       strings.TrimSpace(msg.Content),
		ActionTime: createdAt,
	}
	_ = merchantId
	_ = agentId
	if msg.Sender != nil {
		typing.SenderType = chat.ChatSenderType(msg.Sender.Type)
	}
	return typing
}

func receiptFromMessage(merchantId int64, msg *models.ChatMessage, session *models.TChatSession, createdAt int64) *chat.ChatMessageReceiptPayload {
	if msg == nil {
		return nil
	}
	if session != nil && merchantId <= 0 {
		merchantId = session.MerchantId
	}
	receipt := &chat.ChatMessageReceiptPayload{
		SessionNo:   msg.SessionNo,
		MessageNo:   msg.MessageNo,
		ReceiptTime: createdAt,
	}
	if msg.Sender != nil {
		receipt.OperatorType = chat.ChatSenderType(msg.Sender.Type)
		receipt.OperatorId = msg.Sender.Id
	}
	_ = merchantId
	if session != nil {
		receipt.SenderId = session.UserId
	}
	if sender := msg.Sender; sender != nil {
		receipt.SenderId = sender.Id
	}
	return receipt
}

func messageOperatePayload(req PublishMessageEventReq, createdAt int64) *chat.ChatMessageOperatePayload {
	// msg := eventMessage(req)
	payload := &chat.ChatMessageOperatePayload{
		SessionNo:    req.SessionNo,
		OperatorId:   req.UserId,
		OperatorType: chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
		OperatedAt:   createdAt,
	}
	if req.Message != nil {
		payload.MessageNo = req.Message.MessageNo
		sender := req.Message.Sender
		if sender != nil {
			payload.OperatorId = sender.Id
			payload.OperatorType = chat.ChatSenderType(sender.Type)
		}
	}
	if req.EventType == chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL {
		payload.OperateType = chat.ChatMessageOperateType_CHAT_MESSAGE_OPERATE_TYPE_RECALL
	} else {
		payload.OperateType = chat.ChatMessageOperateType_CHAT_MESSAGE_OPERATE_TYPE_DELETE
	}
	return payload
}
