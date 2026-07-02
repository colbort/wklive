package helper

import (
	"context"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/protobuf/encoding/protojson"
)

func PublishMessageEvent[P any](ctx context.Context, busRedis *redis.Redis, channel string, eventType publishEventType[P], payload P) error {
	createdAt := utils.NowMillis()

	event := &chat.ChatWsResponse{
		Code:      200,
		Msg:       "",
		EventType: eventType.eventType,
		CreatedAt: createdAt,
	}

	if err := setPayload(event, payload); err != nil {
		return err
	}

	payloadBytes, err := protojson.MarshalOptions{UseProtoNames: false, UseEnumNumbers: true}.Marshal(event)
	if err != nil {
		return err
	}

	if _, err := busRedis.PublishCtx(ctx, channel, string(payloadBytes)); err != nil {
		return err
	}

	return nil
}

func setPayload(event *chat.ChatWsResponse, payload any) error {
	switch p := payload.(type) {
	case *chat.ChatWsResponse_Connected:
		event.Payload = p
	case *chat.ChatWsResponse_Message:
		event.Payload = p
	case *chat.ChatWsResponse_SystemNotice:
		event.Payload = p
	case *chat.ChatWsResponse_UserState:
		event.Payload = p
	case *chat.ChatWsResponse_Queue:
		event.Payload = p
	case *chat.ChatWsResponse_Agent:
		event.Payload = p
	case *chat.ChatWsResponse_Transfer:
		event.Payload = p
	case *chat.ChatWsResponse_Session:
		event.Payload = p
	case *chat.ChatWsResponse_Evaluation:
		event.Payload = p
	case *chat.ChatWsResponse_Typing:
		event.Payload = p
	case *chat.ChatWsResponse_Receipt:
		event.Payload = p
	case *chat.ChatWsResponse_MessageOperate:
		event.Payload = p
	case *chat.ChatWsResponse_Heartbeat:
		event.Payload = p
	case *chat.ChatWsResponse_Error:
		event.Payload = p
	default:
		return fmt.Errorf("unsupported payload type: %T", payload)
	}
	return nil
}

type publishEventType[P any] struct {
	eventType chat.ChatEventType
}

var (
	PublishEventWSConnected      = publishEventType[*chat.ChatWsResponse_Connected]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_WS_CONNECTED}
	PublishEventMessage          = publishEventType[*chat.ChatWsResponse_Message]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE}
	PublishEventSystemNotice     = publishEventType[*chat.ChatWsResponse_SystemNotice]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM_NOTICE}
	PublishEventUserJoin         = publishEventType[*chat.ChatWsResponse_UserState]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN}
	PublishEventUserLeave        = publishEventType[*chat.ChatWsResponse_UserState]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE}
	PublishEventQueueUpdate      = publishEventType[*chat.ChatWsResponse_Queue]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE}
	PublishEventAgentJoin        = publishEventType[*chat.ChatWsResponse_Agent]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_JOIN}
	PublishEventAgentAccepted    = publishEventType[*chat.ChatWsResponse_Agent]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ACCEPTED}
	PublishEventAgentLeave       = publishEventType[*chat.ChatWsResponse_Agent]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_LEAVE}
	PublishEventTransferRequest  = publishEventType[*chat.ChatWsResponse_Transfer]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REQUEST}
	PublishEventTransferAccept   = publishEventType[*chat.ChatWsResponse_Transfer]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_ACCEPT}
	PublishEventTransferReject   = publishEventType[*chat.ChatWsResponse_Transfer]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REJECT}
	PublishEventSessionClose     = publishEventType[*chat.ChatWsResponse_Session]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE}
	PublishEventEvaluationInvite = publishEventType[*chat.ChatWsResponse_Evaluation]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE}
	PublishEventEvaluationSubmit = publishEventType[*chat.ChatWsResponse_Evaluation]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT}
	PublishEventTyping           = publishEventType[*chat.ChatWsResponse_Typing]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_TYPING}
	PublishEventMessageDelivered = publishEventType[*chat.ChatWsResponse_Receipt]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELIVERED}
	PublishEventMessageRead      = publishEventType[*chat.ChatWsResponse_Receipt]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_READ}
	PublishEventMessageRecall    = publishEventType[*chat.ChatWsResponse_MessageOperate]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL}
	PublishEventMessageDelete    = publishEventType[*chat.ChatWsResponse_MessageOperate]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELETE}
	PublishEventHeartbeat        = publishEventType[*chat.ChatWsResponse_Heartbeat]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT}
	PublishEventError            = publishEventType[*chat.ChatWsResponse_Error]{eventType: chat.ChatEventType_CHAT_EVENT_TYPE_ERROR}
)
