package internal

import (
	"context"
	"fmt"
	"strconv"
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
	case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Message{Message: msg}
	case chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM_NOTICE:
		event.Payload = &chat.ChatMessageEvent_SystemNotice{SystemNotice: systemNoticePayload(req)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE:
		event.Payload = &chat.ChatMessageEvent_Queue{Queue: queuePayload(req, createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ACCEPTED,
		chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_LEAVE:
		event.Payload = &chat.ChatMessageEvent_Agent{Agent: agentPayload(req, createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Evaluation{Evaluation: evaluationInviteFromMessage(req.MerchantId, msg, eventSession(ctx, svcCtx, req), createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT:
		if req.Satisfaction == nil {
			return fmt.Errorf("satisfaction data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Evaluation{Evaluation: evaluationSubmitPayload(req.Satisfaction)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Typing{Typing: typingFromMessage(req.MerchantId, msg, eventSession(ctx, svcCtx, req), createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELIVERED,
		chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_READ:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Receipt{Receipt: receiptFromMessage(req.MerchantId, msg, eventSession(ctx, svcCtx, req), createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL,
		chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELETE:
		event.Payload = &chat.ChatMessageEvent_MessageOperate{MessageOperate: messageOperatePayload(req, createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT:
		event.Payload = &chat.ChatMessageEvent_Heartbeat{Heartbeat: &chat.ChatHeartbeatPayload{ServerTime: createdAt}}
	case chat.ChatEventType_CHAT_EVENT_TYPE_ERROR:
		msg := eventMessage(req)
		if msg == nil {
			return fmt.Errorf("message data is empty")
		}
		event.Payload = &chat.ChatMessageEvent_Error{Error: chatEventErrorFromMessage(req.MerchantId, msg, eventSession(ctx, svcCtx, req), createdAt)}
	case chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN,
		chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE,
		chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REQUEST,
		chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_ACCEPT,
		chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REJECT,
		chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE:
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
		fmt.Printf("=========== 11  %v  %v\n", req.Channel, req.EventType)
		return err
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, req.Channel, string(payload)); err != nil {
		fmt.Printf("=========== 22  %v  %v\n", req.Channel, req.EventType)
		return err
	}
	fmt.Printf("=========== 33  %v  %v\n", req.Channel, req.EventType)
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

func systemNoticePayload(req PublishMessageEventReq) *chat.ChatSystemNoticePayload {
	sessionNo := strings.TrimSpace(req.SessionNo)
	if req.TransientSession != nil && sessionNo == "" {
		sessionNo = req.TransientSession.GetSessionNo()
	}
	if req.Session != nil && sessionNo == "" {
		sessionNo = req.Session.SessionNo
	}
	return &chat.ChatSystemNoticePayload{
		SessionNo:  sessionNo,
		Content:    strings.TrimSpace(req.EventMessage),
		Level:      "info",
		ShowInChat: false,
	}
}

func queuePayload(req PublishMessageEventReq, createdAt int64) *chat.ChatQueuePayload {
	session := req.TransientSession
	if session == nil && req.Session != nil {
		session = ToProtoSession(req.Session, req.IsGuest)
	}
	sessionNo := strings.TrimSpace(req.SessionNo)
	userId := req.UserId
	status := chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING
	if session != nil {
		if sessionNo == "" {
			sessionNo = session.GetSessionNo()
		}
		if userId <= 0 {
			userId = session.GetUserId()
		}
		status = session.GetStatus()
	}
	return &chat.ChatQueuePayload{
		SessionNo:     sessionNo,
		UserId:        userId,
		QueueAction:   chat.ChatQueueAction_CHAT_QUEUE_ACTION_UPDATE,
		SessionStatus: status,
		ActionTime:    createdAt,
	}
}

func agentPayload(req PublishMessageEventReq, createdAt int64) *chat.ChatAgentPayload {
	session := req.TransientSession
	if session == nil && req.Session != nil {
		session = ToProtoSession(req.Session, req.IsGuest)
	}
	sessionNo := strings.TrimSpace(req.SessionNo)
	agentId := req.AgentId
	status := chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN
	if session != nil {
		if sessionNo == "" {
			sessionNo = session.GetSessionNo()
		}
		if agentId <= 0 {
			agentId = session.GetAgentId()
		}
		status = session.GetStatus()
	}
	payload := &chat.ChatAgentPayload{
		SessionNo:     sessionNo,
		AgentId:       strconv.FormatInt(agentId, 10),
		AssignType:    req.AssignType,
		SessionStatus: status,
		Remark:        strings.TrimSpace(req.EventMessage),
		ActionTime:    createdAt,
	}
	if req.Agent != nil {
		payload.AgentId = strconv.FormatInt(req.Agent.Id, 10)
		payload.AgentStatus = chat.ChatAgentStatus(req.Agent.Status)
	}
	return payload
}

func chatEventErrorFromMessage(merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatErrorPayload {
	if session != nil && merchantId <= 0 {
		merchantId = session.GetMerchantId()
	}
	_ = merchantId
	_ = createdAt
	return &chat.ChatErrorPayload{
		SessionNo:    msg.GetSessionNo(),
		ErrorCode:    500,
		ErrorMessage: strings.TrimSpace(msg.GetContent()),
		Retryable:    false,
	}
}

func evaluationInviteFromMessage(merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatEvaluationPayload {
	if msg == nil {
		return nil
	}
	if session != nil && merchantId <= 0 {
		merchantId = session.GetMerchantId()
	}
	invite := &chat.ChatEvaluationPayload{
		SessionNo:   msg.GetSessionNo(),
		Comment:     strings.TrimSpace(msg.GetContent()),
		Submitted:   false,
		EvaluatedAt: createdAt,
	}
	_ = merchantId
	if session != nil {
		invite.UserId = strconv.FormatInt(session.GetUserId(), 10)
		invite.AgentId = strconv.FormatInt(session.GetAgentId(), 10)
	}
	if sender := msg.GetSender(); sender != nil && sender.GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		invite.AgentId = strconv.FormatInt(sender.GetId(), 10)
	}
	if receiver := msg.GetReceiver(); receiver != nil && receiver.GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
		invite.UserId = strconv.FormatInt(receiver.GetId(), 10)
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
		UserId:       strconv.FormatInt(data.UserId, 10),
		AgentId:      strconv.FormatInt(data.AgentId, 10),
		EvaluationId: strconv.FormatInt(data.Id, 10),
		Rating:       int32(data.Score),
		Tags:         tags,
		Comment:      data.Content,
		Submitted:    true,
		EvaluatedAt:  data.UpdateTimes,
	}
}

func typingFromMessage(merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatTypingPayload {
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
	typing := &chat.ChatTypingPayload{
		SessionNo:  msg.GetSessionNo(),
		SenderId:   strconv.FormatInt(userId, 10),
		Text:       strings.TrimSpace(msg.GetContent()),
		ActionTime: createdAt,
	}
	_ = merchantId
	_ = agentId
	if msg.GetSender() != nil {
		typing.SenderType = msg.GetSender().GetType()
	}
	return typing
}

func receiptFromMessage(merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, createdAt int64) *chat.ChatMessageReceiptPayload {
	if msg == nil {
		return nil
	}
	if session != nil && merchantId <= 0 {
		merchantId = session.GetMerchantId()
	}
	receipt := &chat.ChatMessageReceiptPayload{
		SessionNo:   msg.GetSessionNo(),
		MessageNo:   msg.GetMessageNo(),
		ReceiptTime: createdAt,
	}
	if msg.GetSender() != nil {
		receipt.OperatorType = msg.GetSender().GetType()
		receipt.OperatorId = msg.GetSender().GetId()
	}
	_ = merchantId
	if session != nil {
		receipt.SenderId = session.GetUserId()
	}
	if sender := msg.GetSender(); sender != nil {
		receipt.SenderId = sender.GetId()
	}
	return receipt
}

func messageOperatePayload(req PublishMessageEventReq, createdAt int64) *chat.ChatMessageOperatePayload {
	msg := eventMessage(req)
	payload := &chat.ChatMessageOperatePayload{
		SessionNo:    strings.TrimSpace(req.SessionNo),
		OperatorId:   strconv.FormatInt(req.UserId, 10),
		OperatorType: chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
		OperatedAt:   createdAt,
	}
	if msg != nil {
		payload.SessionNo = msg.GetSessionNo()
		payload.MessageNo = msg.GetMessageNo()
		if sender := msg.GetSender(); sender != nil {
			payload.OperatorId = strconv.FormatInt(sender.GetId(), 10)
			payload.OperatorType = sender.GetType()
		}
	}
	if req.EventType == chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL {
		payload.OperateType = chat.ChatMessageOperateType_CHAT_MESSAGE_OPERATE_TYPE_RECALL
	} else {
		payload.OperateType = chat.ChatMessageOperateType_CHAT_MESSAGE_OPERATE_TYPE_DELETE
	}
	return payload
}
