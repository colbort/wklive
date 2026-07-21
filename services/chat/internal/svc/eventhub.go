package svc

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"wklive/common/mq/kafka"
	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/logx"
)

const chatEventBuffer = 256

type ChatEventHub struct {
	mu          sync.Mutex
	nextID      atomic.Uint64
	subscribers map[string]map[uint64]chan []byte
}

func NewChatEventHub() *ChatEventHub {
	return &ChatEventHub{subscribers: make(map[string]map[uint64]chan []byte)}
}

func (h *ChatEventHub) Subscribe(channel string) (<-chan []byte, func()) {
	id := h.nextID.Add(1)
	ch := make(chan []byte, chatEventBuffer)
	h.mu.Lock()
	if h.subscribers[channel] == nil {
		h.subscribers[channel] = make(map[uint64]chan []byte)
	}
	h.subscribers[channel][id] = ch
	h.mu.Unlock()
	return ch, func() { h.remove(channel, id) }
}

func (h *ChatEventHub) Publish(channel string, payload []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	for id, ch := range h.subscribers[channel] {
		copyPayload := append([]byte(nil), payload...)
		select {
		case ch <- copyPayload:
		default:
			close(ch)
			delete(h.subscribers[channel], id)
			logx.Errorf("disconnect slow chat event subscriber, channel=%s id=%d", channel, id)
		}
	}
	return nil
}

func (h *ChatEventHub) remove(channel string, id uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.subscribers[channel][id]; ok {
		close(ch)
		delete(h.subscribers[channel], id)
	}
}

func (h *ChatEventHub) Start(ctx context.Context, config mq.Config) {
	instance := hostName()
	for _, channel := range []string{chat.ChatAppEventChannel, chat.ChatAdminEventChannel} {
		channel := channel
		subscriber := mq.MustNewBroadcastSubscriber(config, "chat-events-"+mq.Topic(channel)+"-"+instance)
		go func() {
			for {
				err := subscriber.Subscribe(ctx, channel, func(_ context.Context, message mq.Message) error {
					return h.Publish(channel, []byte(message.Payload))
				})
				if ctx.Err() != nil {
					return
				}
				if err != nil && !errors.Is(err, context.Canceled) {
					logx.Errorf("chat mq subscription failed, channel=%s err=%v", channel, err)
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(3 * time.Second):
				}
			}
		}()
	}
}

func hostName() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "unknown"
	}
	return name
}
