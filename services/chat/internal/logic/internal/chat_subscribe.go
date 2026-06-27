package internal

import (
	"context"
	"fmt"
	"strings"

	v9 "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/protobuf/encoding/protojson"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
)

type ChatEventStream interface {
	Context() context.Context
	Send(*chat.ChatMessageEvent) error
}

type ChatSubscribeRequest interface {
	GetMerchantId() int64
	GetUserId() int64
	GetAgentId() int64
	GetSessionNo() string
	GetIsGuest() bool
	GetAdmin() bool
}

type SubscribeOptions struct {
	Channel string
	Admin   bool
}

type redisSubscriber interface {
	Subscribe(ctx context.Context, channels ...string) *v9.PubSub
}

func SubscribeChatEventStream(svcCtx *svc.ServiceContext, req ChatSubscribeRequest, stream ChatEventStream, opts SubscribeOptions) error {
	if svcCtx == nil || svcCtx.BusRedis == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	rds, err := redis.NewRedis(svcCtx.Config.Redis.RedisConf)
	if err != nil {
		return err
	}

	node, err := redis.CreateBlockingNode(rds)
	if err != nil {
		return err
	}
	defer node.Close()

	client, ok := node.(redisSubscriber)
	if !ok {
		return fmt.Errorf("redis node does not support subscribe")
	}

	ctx := stream.Context()
	pubsub := client.Subscribe(ctx, opts.Channel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			var event chat.ChatMessageEvent
			if err := protojson.Unmarshal([]byte(msg.Payload), &event); err != nil {
				logx.WithContext(ctx).Errorf("decode chat stream event failed: %v", err)
				continue
			}
			if !matchSubscribeEvent(req, &event, opts.Admin) {
				continue
			}
			if err := stream.Send(&event); err != nil {
				return err
			}
		}
	}
}

func matchSubscribeEvent(req ChatSubscribeRequest, event *chat.ChatMessageEvent, admin bool) bool {
	if req == nil || event == nil {
		return false
	}
	if !admin && !shouldBroadcastToUser(event) {
		return false
	}
	if event.GetMessage() != nil && matchSubscribeMessage(req, event.GetMessage(), admin) {
		return true
	}
	if event.GetSession() != nil && matchSubscribeSession(req, event.GetSession(), admin) {
		return true
	}
	if event.GetQueue() != nil && matchSubscribeQueue(req, event.GetQueue(), admin) {
		return true
	}
	if event.GetSatisfaction() != nil && matchSubscribeSatisfaction(req, event.GetSatisfaction(), admin) {
		return true
	}
	if admin && event.GetAgent() != nil && matchSubscribeAgent(req, event.GetAgent()) {
		return true
	}
	return false
}

func shouldBroadcastToUser(event *chat.ChatMessageEvent) bool {
	switch event.GetEventType() {
	case chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN,
		chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE:
		return false
	default:
		return true
	}
}

func matchSubscribeMessage(req ChatSubscribeRequest, msg *chat.ChatMessage, admin bool) bool {
	if msg == nil {
		return false
	}
	if !admin {
		if strings.TrimSpace(req.GetSessionNo()) != "" {
			return msg.GetSessionNo() == strings.TrimSpace(req.GetSessionNo())
		}
		return matchMessageUser(req.GetUserId(), msg)
	}
	if strings.TrimSpace(req.GetSessionNo()) != "" && msg.GetSessionNo() != strings.TrimSpace(req.GetSessionNo()) {
		return false
	}
	agentId := messageAgentID(msg)
	if req.GetAgentId() > 0 && agentId != req.GetAgentId() && agentId != 0 {
		return false
	}
	return true
}

func matchMessageUser(userId int64, msg *chat.ChatMessage) bool {
	if userId <= 0 || msg == nil {
		return false
	}
	if msg.GetSender() != nil && msg.GetSender().GetId() == userId {
		return true
	}
	if msg.GetReceiver() != nil && msg.GetReceiver().GetId() == userId {
		return true
	}
	return false
}

func matchSubscribeSession(req ChatSubscribeRequest, session *chat.ChatSession, admin bool) bool {
	if session == nil {
		return false
	}
	if req.GetMerchantId() > 0 && session.GetMerchantId() != req.GetMerchantId() {
		return false
	}
	if strings.TrimSpace(req.GetSessionNo()) != "" && session.GetSessionNo() != strings.TrimSpace(req.GetSessionNo()) {
		return false
	}
	if !admin {
		if strings.TrimSpace(req.GetSessionNo()) == "" && req.GetUserId() > 0 && session.GetUserId() != req.GetUserId() {
			return false
		}
		return true
	}
	if req.GetAgentId() > 0 && session.GetAgentId() != req.GetAgentId() && session.GetAgentId() != 0 {
		return false
	}
	return true
}

func matchSubscribeQueue(req ChatSubscribeRequest, queue *chat.ChatQueueInfo, admin bool) bool {
	if queue == nil {
		return false
	}
	if req.GetMerchantId() > 0 && queue.GetMerchantId() != req.GetMerchantId() {
		return false
	}
	if strings.TrimSpace(req.GetSessionNo()) != "" && queue.GetSessionNo() != strings.TrimSpace(req.GetSessionNo()) {
		return false
	}
	if !admin && strings.TrimSpace(req.GetSessionNo()) == "" && req.GetUserId() > 0 && queue.GetUserId() != req.GetUserId() {
		return false
	}
	return true
}

func matchSubscribeSatisfaction(req ChatSubscribeRequest, satisfaction *chat.ChatSatisfaction, admin bool) bool {
	if satisfaction == nil {
		return false
	}
	if req.GetMerchantId() > 0 && satisfaction.GetMerchantId() != req.GetMerchantId() {
		return false
	}
	if strings.TrimSpace(req.GetSessionNo()) != "" && satisfaction.GetSessionNo() != strings.TrimSpace(req.GetSessionNo()) {
		return false
	}
	if !admin && strings.TrimSpace(req.GetSessionNo()) == "" && req.GetUserId() > 0 && satisfaction.GetUserId() != req.GetUserId() {
		return false
	}
	if admin && req.GetAgentId() > 0 && satisfaction.GetAgentId() != req.GetAgentId() && satisfaction.GetAgentId() != 0 {
		return false
	}
	return true
}

func matchSubscribeAgent(req ChatSubscribeRequest, agent *chat.ChatAgent) bool {
	if agent == nil {
		return false
	}
	if req.GetMerchantId() > 0 && agent.GetMerchantId() != req.GetMerchantId() {
		return false
	}
	if req.GetAgentId() > 0 && agent.GetId() != req.GetAgentId() {
		return false
	}
	return true
}

func messageAgentID(msg *chat.ChatMessage) int64 {
	if msg == nil || msg.GetSender() == nil {
		return 0
	}
	if msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		return msg.GetSender().GetId()
	}
	return 0
}
