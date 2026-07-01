package helper

import (
	"context"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/protobuf/encoding/protojson"
)

type PublishMessageEventReq struct {
	Ctx       context.Context
	BusRedis  *redis.Redis
	Channel   string
	EventType chat.ChatEventType
	Payload   any
}

func PublishMessageEvent(req PublishMessageEventReq) error {
	createdAt := utils.NowMillis()

	event := &chat.ChatMessageEvent{
		Code:      200,
		Msg:       "",
		EventType: req.EventType,
		CreatedAt: createdAt,
	}

	switch p := req.Payload.(type) {
	case *chat.ChatMessageEvent_Message:
		event.Payload = p
	case *chat.ChatMessageEvent_Session:
		event.Payload = p
	case *chat.ChatMessageEvent_SystemNotice:
		event.Payload = p
	case *chat.ChatMessageEvent_UserState:
		event.Payload = p
	case *chat.ChatMessageEvent_Queue:
		event.Payload = p
	case *chat.ChatMessageEvent_Agent:
		event.Payload = p
	case *chat.ChatMessageEvent_Transfer:
		event.Payload = p
	case *chat.ChatMessageEvent_Evaluation:
		event.Payload = p
	case *chat.ChatMessageEvent_Typing:
		event.Payload = p
	case *chat.ChatMessageEvent_Receipt:
		event.Payload = p
	case *chat.ChatMessageEvent_MessageOperate:
		event.Payload = p
	case *chat.ChatMessageEvent_Heartbeat:
		event.Payload = p
	case *chat.ChatMessageEvent_Error:
		event.Payload = p

	default:
		return fmt.Errorf("unsupported payload type: %T", req.Payload)
	}

	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		return err
	}

	if _, err := req.BusRedis.PublishCtx(req.Ctx, req.Channel, string(payload)); err != nil {
		return err
	}

	return nil
}
