package server

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/zeromicro/go-zero/core/logx"
)

var globalSubscriberID int64

type Subscriber struct {
	ID int64
	ch chan ServerMessage
}

func NewSubscriber(buf int) *Subscriber {
	if buf <= 0 {
		buf = 256
	}

	return &Subscriber{
		ID: atomic.AddInt64(&globalSubscriberID, 1),
		ch: make(chan ServerMessage, buf),
	}
}

func (s *Subscriber) C() <-chan ServerMessage {
	return s.ch
}

type FirstSubscribeHook func(key string, msg ClientMessage)
type LastUnsubscribeHook func(key string, msg ClientMessage)

type Hub struct {
	mu sync.RWMutex

	// topic -> subscribers
	subs map[string]map[*Subscriber]struct{}

	// subscriber -> topic keys
	subTopics map[*Subscriber]map[string]struct{}

	// topic -> 是否已经向上游订阅
	upstreamSubscribed map[string]bool

	onFirstSubscribe FirstSubscribeHook
	onLastLeave      LastUnsubscribeHook
}

func NewHub() *Hub {
	return &Hub{
		subs:               make(map[string]map[*Subscriber]struct{}),
		subTopics:          make(map[*Subscriber]map[string]struct{}),
		upstreamSubscribed: make(map[string]bool),
	}
}

func (h *Hub) SetHooks(onFirst FirstSubscribeHook, onLast LastUnsubscribeHook) {
	h.onFirstSubscribe = onFirst
	h.onLastLeave = onLast
}

func (h *Hub) NewSubscriber(buf int) *Subscriber {
	sub := NewSubscriber(buf)

	h.mu.Lock()
	h.subTopics[sub] = make(map[string]struct{})
	h.mu.Unlock()

	return sub
}

func (h *Hub) Subscribe(sub *Subscriber, msg ClientMessage) error {
	msg = NormalizeClientMessage(msg)

	if sub == nil {
		return errors.New("subscriber is nil")
	}
	if msg.Topic == "" || msg.CategoryCode == "" || msg.Symbol == "" || msg.Market == "" {
		return errors.New("invalid subscribe message")
	}
	if msg.Topic == TopicKline && msg.Interval == "" {
		return errors.New("interval is required for kline")
	}

	key := BuildTopicKey(msg)
	needCallHook := false

	h.mu.Lock()

	set, ok := h.subs[key]
	if !ok {
		set = make(map[*Subscriber]struct{})
		h.subs[key] = set
	}

	if _, existed := set[sub]; existed {
		h.mu.Unlock()
		return nil
	}

	set[sub] = struct{}{}

	if _, ok := h.subTopics[sub]; !ok {
		h.subTopics[sub] = make(map[string]struct{})
	}
	h.subTopics[sub][key] = struct{}{}

	if !h.upstreamSubscribed[key] {
		h.upstreamSubscribed[key] = true
		needCallHook = true
	}

	h.mu.Unlock()

	logx.Infof("stream subscribe success, subscriber=%d, topic=%s", sub.ID, key)

	if needCallHook && h.onFirstSubscribe != nil {
		go h.onFirstSubscribe(key, msg)
	}

	return nil
}

func (h *Hub) Unsubscribe(sub *Subscriber, msg ClientMessage) error {
	msg = NormalizeClientMessage(msg)

	if sub == nil {
		return errors.New("subscriber is nil")
	}
	if msg.Topic == "" || msg.CategoryCode == "" || msg.Symbol == "" || msg.Market == "" {
		return errors.New("invalid unsubscribe message")
	}

	key := BuildTopicKey(msg)
	return h.unsubscribeByKey(sub, key, msg)
}

func (h *Hub) unsubscribeByKey(sub *Subscriber, key string, msg ClientMessage) error {
	needCallHook := false

	h.mu.Lock()

	set, ok := h.subs[key]
	if !ok {
		h.mu.Unlock()
		return nil
	}

	if _, existed := set[sub]; !existed {
		h.mu.Unlock()
		return nil
	}

	delete(set, sub)

	if topics, ok := h.subTopics[sub]; ok {
		delete(topics, key)
	}

	if len(set) == 0 {
		delete(h.subs, key)
		delete(h.upstreamSubscribed, key)
		needCallHook = true
	}

	h.mu.Unlock()

	logx.Infof("stream unsubscribe success, subscriber=%d, topic=%s", sub.ID, key)

	if needCallHook && h.onLastLeave != nil {
		go h.onLastLeave(key, msg)
	}

	return nil
}

func (h *Hub) RemoveSubscriber(sub *Subscriber) {
	if sub == nil {
		return
	}

	h.mu.Lock()
	topicsMap := h.subTopics[sub]
	topics := make([]string, 0, len(topicsMap))
	for topic := range topicsMap {
		topics = append(topics, topic)
	}
	h.mu.Unlock()

	for _, topic := range topics {
		msg := ParseTopicKey(topic)
		_ = h.unsubscribeByKey(sub, topic, msg)
	}

	h.mu.Lock()
	delete(h.subTopics, sub)
	h.mu.Unlock()

	close(sub.ch)

	logx.Infof("stream subscriber removed, subscriber=%d", sub.ID)
}

func (h *Hub) Broadcast(topicMsg ClientMessage, payload any) {
	topicMsg = NormalizeClientMessage(topicMsg)
	key := BuildTopicKey(topicMsg)

	msg := ServerMessage{
		Topic:        topicMsg.Topic,
		CategoryCode: topicMsg.CategoryCode,
		Symbol:       topicMsg.Symbol,
		Market:       topicMsg.Market,
		Interval:     topicMsg.Interval,
		Payload:      payload,
	}

	h.mu.RLock()
	set := h.subs[key]
	subs := make([]*Subscriber, 0, len(set))
	for sub := range set {
		subs = append(subs, sub)
	}
	h.mu.RUnlock()

	if len(subs) == 0 {
		logx.Errorf("stream broadcast skipped, no local subscriber, topic=%s", key)
		return
	}

	for _, sub := range subs {
		select {
		case sub.ch <- msg:
		default:
			logx.Errorf("stream subscriber queue full, subscriber=%d, topic=%s", sub.ID, key)
		}
	}
}

func (h *Hub) TopicSubscriberCount(msg ClientMessage) int {
	key := BuildTopicKey(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.subs[key])
}

func (h *Hub) SnapshotSubscriptions(categoryCode string) []ClientMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()

	categoryCode = strings.ToLower(strings.TrimSpace(categoryCode))
	out := make([]ClientMessage, 0, len(h.subs))

	for key, set := range h.subs {
		if len(set) == 0 {
			continue
		}

		msg := ParseTopicKey(key)
		if strings.ToLower(strings.TrimSpace(msg.CategoryCode)) != categoryCode {
			continue
		}

		out = append(out, msg)
	}

	return out
}
