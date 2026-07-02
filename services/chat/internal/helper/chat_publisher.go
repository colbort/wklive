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

	event := &chat.ChatWsResponse{
		Code:      200,
		Msg:       "",
		EventType: req.EventType,
		CreatedAt: createdAt,
	}

	switch p := req.Payload.(type) {
	case *chat.ChatWsResponse_Message:
		event.Payload = p
	case *chat.ChatWsResponse_Session:
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
