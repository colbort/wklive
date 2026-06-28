package internal

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
	EventType        chat.ChatEventType
	Channel          string
	MerchantId       int64
	SessionNo        string
	UserId           int64
	AgentId          int64
	IsGuest          bool
	AssignType       chat.ChatAssignType
	Reason           string
	EventMessage     string
	Message          *models.ChatMessage
	TransientMessage *chat.ChatMessage
	Session          *models.TChatSession
	TransientSession *chat.ChatSession
	Agent            *models.TChatAgent
	Satisfaction     *models.TChatSatisfaction
}

func PublishMessageEvent(ctx context.Context, svcCtx *svc.ServiceContext, req PublishMessageEventReq) error {
	if req.EventType == chat.ChatEventType_CHAT_EVENT_TYPE_UNSPECIFIED {
		return fmt.Errorf("event_type is required")
	}
	createdAt := utils.NowMillis()
	event := &chat.ChatMessageEvent{EventType: req.EventType, CreatedAt: createdAt}
	switch req.EventType {
	case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Message{Message: msg}
	case chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_JOIN,
		chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE:
		event.Payload = &chat.ChatMessageEvent_Queue{Queue: &chat.ChatQueueInfo{
			MerchantId: req.MerchantId,
			SessionNo:  req.SessionNo,
			Message:    req.EventMessage,
		}}
	case chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_JOIN,
		chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_LEAVE:
		if req.Agent == nil {
			return fmt.Errorf("agent data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Agent{Agent: ToProtoAgent(req.Agent)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_EvaluationInvite{EvaluationInvite: evaluationInviteFromMessage(req.MerchantId, msg, eventSession(ctx, svcCtx, req), createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT:
		if req.Satisfaction == nil {
			return fmt.Errorf("satisfaction data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Satisfaction{Satisfaction: ToProtoSatisfaction(req.Satisfaction)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING,
		chat.ChatEventType_CHAT_EVENT_TYPE_STOP_TYPING:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Typing{Typing: typingFromMessage(req.MerchantId, msg, eventSession(ctx, svcCtx, req), createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_DELIVERED,
		chat.ChatEventType_CHAT_EVENT_TYPE_READ,
		chat.ChatEventType_CHAT_EVENT_TYPE_RECALL,
		chat.ChatEventType_CHAT_EVENT_TYPE_DELETE:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Receipt{Receipt: receiptFromMessage(req.MerchantId, msg, eventSession(ctx, svcCtx, req), createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT:
		event.Payload = &chat.ChatMessageEvent_Heartbeat{Heartbeat: &chat.ChatHeartbeat{Time: createdAt}}
	case chat.ChatEventType_CHAT_EVENT_TYPE_ERROR:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Error{Error: chatEventErrorFromMessage(req.MerchantId, msg, eventSession(ctx, svcCtx, req), createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN,
		chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE,
		chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED,
		chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REQUEST,
		chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_ACCEPT,
		chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REJECT,
		chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_START,
		chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE,
		chat.ChatEventType_CHAT_EVENT_TYPE_NO_AGENT_ONLINE,
		chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_TIMEOUT:
		session, err := eventSessionPayload(req, createdAt)
		if err != nil {
			return err
		}
		event.Payload = &chat.ChatMessageEvent_Session{Session: session}
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

func eventMessage(req PublishMessageEventReq) *chat.ChatMessage {
	if req.TransientMessage != nil {
		return req.TransientMessage
	}
	return ToProtoMessage(req.Message)
}

func eventSession(ctx context.Context, svcCtx *svc.ServiceContext, req PublishMessageEventReq) *chat.ChatSession {
	if req.TransientSession != nil {
		return req.TransientSession
	}
	if req.TransientMessage != nil && svcCtx != nil && svcCtx.BusRedis != nil {
		session, _ := GetTransientSession(ctx, svcCtx.BusRedis, req.MerchantId, req.TransientMessage.GetSessionNo())
		return session
	}
	if req.Session != nil {
		return ToProtoSession(req.Session, req.IsGuest)
	}
	return nil
}

func eventSessionPayload(req PublishMessageEventReq, createdAt int64) (*chat.ChatSession, error) {
	if req.Session != nil {
		return sessionWithEventMeta(ToProtoSession(req.Session, req.IsGuest), req.EventType, strings.TrimSpace(req.EventMessage), strings.TrimSpace(req.Reason), req.AssignType, createdAt), nil
	}
	session := req.TransientSession
	sessionNo := strings.TrimSpace(req.SessionNo)
	if session != nil && sessionNo == "" {
		sessionNo = session.GetSessionNo()
	}
	if strings.TrimSpace(sessionNo) == "" {
		return nil, fmt.Errorf("session_no is required")
	}
	merchantId := req.MerchantId
	userId := req.UserId
	agentId := req.AgentId
	if session != nil {
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
	if session == nil {
		session = &chat.ChatSession{SessionNo: sessionNo, MerchantId: merchantId, UserId: userId, AgentId: agentId}
	}
	return sessionWithEventMeta(session, req.EventType, strings.TrimSpace(req.EventMessage), strings.TrimSpace(req.Reason), req.AssignType, createdAt), nil
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

func chatEventErrorFromMessage(merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatEventError {
	if session != nil && merchantId <= 0 {
		merchantId = session.GetMerchantId()
	}
	return &chat.ChatEventError{
		Code:       500,
		Msg:        strings.TrimSpace(msg.GetContent()),
		SessionNo:  msg.GetSessionNo(),
		MerchantId: merchantId,
		CreatedAt:  createdAt,
	}
}

func evaluationInviteFromMessage(merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatEvaluationInvite {
	if msg == nil {
		return nil
	}
	if session != nil && merchantId <= 0 {
		merchantId = session.GetMerchantId()
	}
	invite := &chat.ChatEvaluationInvite{
		SessionNo:  msg.GetSessionNo(),
		MerchantId: merchantId,
		Content:    strings.TrimSpace(msg.GetContent()),
		CreatedAt:  createdAt,
	}
	if session != nil {
		invite.UserId = session.GetUserId()
		invite.AgentId = session.GetAgentId()
	}
	if sender := msg.GetSender(); sender != nil && sender.GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		invite.AgentId = sender.GetId()
	}
	if receiver := msg.GetReceiver(); receiver != nil && receiver.GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
		invite.UserId = receiver.GetId()
	}
	return invite
}

func typingFromMessage(merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatTyping {
	if msg == nil {
		return nil
	}
	if session != nil && merchantId <= 0 {
		merchantId = session.GetMerchantId()
	}
	userId := int64(0)
	agentId := int64(0)
	if session != nil {
		userId = session.GetUserId()
		agentId = session.GetAgentId()
	}
	if msg.GetSender() != nil {
		if msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
			userId = msg.GetSender().GetId()
		}
		if msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
			agentId = msg.GetSender().GetId()
		}
	}
	typing := &chat.ChatTyping{
		SessionNo:  msg.GetSessionNo(),
		MerchantId: merchantId,
		UserId:     userId,
		AgentId:    agentId,
		Message:    strings.TrimSpace(msg.GetContent()),
		CreatedAt:  createdAt,
	}
	if msg.GetSender() != nil {
		typing.SenderType = msg.GetSender().GetType()
	}
	return typing
}

func receiptFromMessage(merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatMessageReceipt {
	if msg == nil {
		return nil
	}
	if session != nil && merchantId <= 0 {
		merchantId = session.GetMerchantId()
	}
	receipt := &chat.ChatMessageReceipt{
		SessionNo:  msg.GetSessionNo(),
		MessageNo:  msg.GetMessageNo(),
		MerchantId: merchantId,
		CreatedAt:  createdAt,
	}
	if msg.GetSender() != nil {
		receipt.SenderType = msg.GetSender().GetType()
	}
	if session != nil {
		receipt.UserId = session.GetUserId()
		receipt.AgentId = session.GetAgentId()
	}
	if sender := msg.GetSender(); sender != nil {
		if sender.GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
			receipt.UserId = sender.GetId()
		}
		if sender.GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
			receipt.AgentId = sender.GetId()
		}
	}
	return receipt
}
