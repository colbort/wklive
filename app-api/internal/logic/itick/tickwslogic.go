// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package itick

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/itick"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type streamReply struct {
	gen   int64
	reply *itick.PushReply
}

type streamError struct {
	gen int64
	err error
}

type tickWsRuntime struct {
	conn *websocket.Conn

	mu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	replyCh     chan streamReply
	writeCh     chan any
	errCh       chan error
	streamErrCh chan streamError

	lastPingAt      int64
	streamGen       int64
	activeStreamGen int64
	streamCancel    context.CancelFunc
	subscriptionMap map[string]types.WsTickTopic
}

type TickWsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTickWsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TickWsLogic {
	return &TickWsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 客户端约定
// 首帧订阅
// {
//   "type": "subscribe",
//   "topics": [
//     {
//       "topic": "depth",    // 参数类型，（depth：盘口、quote：报价、tick：成交、kline：K线）
//       "categoryCode": "crypto",
//       "symbol": "BTCUSDT",
//       "market": "BA",
//       "interval": "1m"    //  kline 订阅必传。（1m，5m，15m，30m，1h，1d，1w，1mo）
//     }
//   ]
// }
// 心跳 ping
// {
//   "type": "ping",
//   "clientTs": 1711888888888
// }
// 服务端 pong
// {
//   "type": "pong",
//   "clientTs": 1711888888888,
//   "serverTs": 1711888889999
// }

func (l *TickWsLogic) TickWs(conn *websocket.Conn, r *http.Request) error {
	defer conn.Close()

	const (
		heartbeatTimeout = 70 * time.Second
		maxPingInterval  = 40 * time.Second
	)

	nowMs := func() int64 {
		return time.Now().UnixMilli()
	}

	if err := conn.SetReadDeadline(time.Now().Add(heartbeatTimeout)); err != nil {
		return err
	}

	if err := conn.WriteJSON(map[string]any{
		"type":     "connected",
		"serverTs": nowMs(),
	}); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(l.ctx)
	defer cancel()

	runtime := &tickWsRuntime{
		conn:            conn,
		ctx:             ctx,
		cancel:          cancel,
		replyCh:         make(chan streamReply, 32),
		writeCh:         make(chan any, 16),
		errCh:           make(chan error, 4),
		streamErrCh:     make(chan streamError, 8),
		subscriptionMap: make(map[string]types.WsTickTopic),
	}

	go l.readLoop(runtime, heartbeatTimeout, maxPingInterval, nowMs)

	return l.writeLoop(runtime, nowMs)
}

func (l *TickWsLogic) readLoop(runtime *tickWsRuntime, heartbeatTimeout, maxPingInterval time.Duration, nowMs func() int64) {
	for {
		var msg types.WsMessage
		if err := runtime.conn.ReadJSON(&msg); err != nil {
			select {
			case runtime.errCh <- err:
			default:
			}
			return
		}

		recvAt := time.Now()

		if err := runtime.conn.SetReadDeadline(recvAt.Add(heartbeatTimeout)); err != nil {
			select {
			case runtime.errCh <- err:
			default:
			}
			return
		}

		switch msg.Type {
		case "ping":
			l.handlePing(runtime, msg, recvAt, maxPingInterval, nowMs)
		case "subscribe":
			if err := l.handleSubscribe(runtime, msg, nowMs); err != nil {
				select {
				case runtime.errCh <- err:
				default:
				}
				return
			}
		case "unsubscribe":
			if err := l.handleUnsubscribe(runtime, msg, nowMs); err != nil {
				select {
				case runtime.errCh <- err:
				default:
				}
				return
			}
		default:
			select {
			case runtime.writeCh <- map[string]any{
				"type":     "error",
				"code":     400,
				"message":  "unsupported message type",
				"serverTs": nowMs(),
			}:
			default:
			}
		}
	}
}

func (l *TickWsLogic) writeLoop(runtime *tickWsRuntime, nowMs func() int64) error {
	for {
		select {
		case <-runtime.ctx.Done():
			return runtime.ctx.Err()

		case err := <-runtime.errCh:
			runtime.cancel()
			runtime.stopStream()
			return err

		case streamErr := <-runtime.streamErrCh:
			if !runtime.isActiveStream(streamErr.gen) {
				continue
			}
			runtime.cancel()
			runtime.stopStream()
			return streamErr.err

		case item := <-runtime.replyCh:
			if !runtime.isActiveStream(item.gen) || item.reply == nil {
				continue
			}

			reply := item.reply
			if err := runtime.conn.WriteJSON(map[string]any{
				"type":         "push",
				"topic":        reply.Topic,
				"categoryCode": reply.CategoryCode,
				"symbol":       reply.Symbol,
				"market":       reply.Market,
				"interval":     reply.Interval,
				"payload":      json.RawMessage(reply.Payload),
				"serverTs":     nowMs(),
			}); err != nil {
				runtime.cancel()
				runtime.stopStream()
				return err
			}

		case out := <-runtime.writeCh:
			if err := runtime.conn.WriteJSON(out); err != nil {
				runtime.cancel()
				runtime.stopStream()
				return err
			}
		}
	}
}

func (l *TickWsLogic) handlePing(runtime *tickWsRuntime, msg types.WsMessage, recvAt time.Time, maxPingInterval time.Duration, nowMs func() int64) {
	currentPingAt := recvAt.UnixMilli()

	if runtime.lastPingAt > 0 {
		interval := time.Duration(currentPingAt-runtime.lastPingAt) * time.Millisecond
		if interval > maxPingInterval {
			select {
			case runtime.writeCh <- map[string]any{
				"type":      "error",
				"code":      4001,
				"message":   "ping interval exceeded",
				"clientTs":  msg.ClientTs,
				"serverTs":  nowMs(),
				"maxMillis": maxPingInterval.Milliseconds(),
				"actualMs":  interval.Milliseconds(),
			}:
			default:
			}

			select {
			case runtime.errCh <- fmt.Errorf("ping interval exceeded: %dms", interval.Milliseconds()):
			default:
			}
			return
		}
	}

	runtime.lastPingAt = currentPingAt

	select {
	case runtime.writeCh <- map[string]any{
		"type":     "pong",
		"clientTs": msg.ClientTs,
		"serverTs": nowMs(),
	}:
	case <-runtime.ctx.Done():
	}
}

func (l *TickWsLogic) handleSubscribe(runtime *tickWsRuntime, msg types.WsMessage, nowMs func() int64) error {
	if len(msg.Topics) == 0 {
		select {
		case runtime.writeCh <- map[string]any{
			"type":     "error",
			"code":     400,
			"message":  "empty subscribe topics",
			"serverTs": nowMs(),
		}:
		default:
		}
		return nil
	}

	added := make([]types.WsTickTopic, 0, len(msg.Topics))
	for _, rawTopic := range msg.Topics {
		topic := normalizeWsTickTopic(rawTopic)
		if !isValidWsTickTopic(topic) {
			continue
		}
		key := buildTopicKey(topic)
		if runtime.hasSubscription(key) {
			continue
		}
		runtime.addSubscription(key, topic)
		added = append(added, topic)
	}

	if len(added) > 0 || (runtime.subscriptionSize() > 0 && !runtime.hasActiveStream()) {
		if err := l.restartSubscriptionStream(runtime); err != nil {
			select {
			case runtime.writeCh <- map[string]any{
				"type":     "error",
				"code":     500,
				"message":  err.Error(),
				"serverTs": nowMs(),
			}:
			default:
			}
			return err
		}
	}

	select {
	case runtime.writeCh <- map[string]any{
		"type":      "subscribed",
		"topics":    msg.Topics,
		"added":     added,
		"serverTs":  nowMs(),
		"topicSize": runtime.subscriptionSize(),
		"current":   runtime.snapshotSubscriptions(),
	}:
	default:
	}

	return nil
}

func (l *TickWsLogic) handleUnsubscribe(runtime *tickWsRuntime, msg types.WsMessage, nowMs func() int64) error {
	if len(msg.Topics) == 0 {
		select {
		case runtime.writeCh <- map[string]any{
			"type":     "error",
			"code":     400,
			"message":  "empty unsubscribe topics",
			"serverTs": nowMs(),
		}:
		default:
		}
		return nil
	}

	removed := make([]string, 0, len(msg.Topics))
	for _, rawTopic := range msg.Topics {
		topic := normalizeWsTickTopic(rawTopic)
		if !isValidWsTickTopic(topic) {
			continue
		}
		key := buildTopicKey(topic)
		if runtime.removeSubscription(key) {
			removed = append(removed, key)
		}
	}

	if len(removed) > 0 || (runtime.subscriptionSize() == 0 && runtime.hasActiveStream()) {
		if err := l.restartSubscriptionStream(runtime); err != nil {
			select {
			case runtime.writeCh <- map[string]any{
				"type":     "error",
				"code":     500,
				"message":  err.Error(),
				"serverTs": nowMs(),
			}:
			default:
			}
			return err
		}
	}

	select {
	case runtime.writeCh <- map[string]any{
		"type":      "unsubscribed",
		"topics":    msg.Topics,
		"removed":   removed,
		"serverTs":  nowMs(),
		"topicSize": runtime.subscriptionSize(),
		"current":   runtime.snapshotSubscriptions(),
	}:
	default:
	}

	return nil
}

func (l *TickWsLogic) restartSubscriptionStream(runtime *tickWsRuntime) error {
	topics := runtime.snapshotSubscriptions()
	runtime.stopStream()

	if len(topics) == 0 {
		return nil
	}

	currentGen := runtime.streamGen + 1

	streamCtx, cancelStream := context.WithCancel(runtime.ctx)
	reqTopics := make([]*itick.SubscribeTopic, 0, len(topics))
	for _, topic := range topics {
		reqTopics = append(reqTopics, &itick.SubscribeTopic{
			Topic:        topic.Topic,
			CategoryCode: topic.CategoryCode,
			Symbol:       topic.Symbol,
			Market:       topic.Market,
			Interval:     topic.Interval,
		})
	}

	stream, err := l.svcCtx.ItickCli.SubscribeStream(streamCtx, &itick.SubscribeRequest{Topics: reqTopics})
	if err != nil {
		cancelStream()
		return err
	}

	runtime.addStream(currentGen, cancelStream)

	go func() {
		for {
			reply, err := stream.Recv()
			if err != nil {
				select {
				case runtime.streamErrCh <- streamError{gen: currentGen, err: err}:
				default:
				}
				return
			}

			select {
			case runtime.replyCh <- streamReply{gen: currentGen, reply: reply}:
			case <-streamCtx.Done():
				return
			}
		}
	}()

	return nil
}

func (r *tickWsRuntime) stopStream() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.streamCancel != nil {
		r.streamCancel()
	}
	r.streamCancel = nil
	r.activeStreamGen = 0
}

func (r *tickWsRuntime) hasSubscription(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.subscriptionMap[key]
	return ok
}

func (r *tickWsRuntime) addSubscription(key string, topic types.WsTickTopic) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subscriptionMap[key] = topic
}

func (r *tickWsRuntime) removeSubscription(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.subscriptionMap[key]; !ok {
		return false
	}
	delete(r.subscriptionMap, key)
	return true
}

func (r *tickWsRuntime) subscriptionSize() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.subscriptionMap)
}

func (r *tickWsRuntime) snapshotSubscriptions() []types.WsTickTopic {
	r.mu.RLock()
	defer r.mu.RUnlock()
	topics := make([]types.WsTickTopic, 0, len(r.subscriptionMap))
	for _, topic := range r.subscriptionMap {
		topics = append(topics, topic)
	}
	return topics
}

func (r *tickWsRuntime) addStream(gen int64, cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.streamGen = gen
	r.activeStreamGen = gen
	r.streamCancel = cancel
}

func (r *tickWsRuntime) isActiveStream(gen int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return gen > 0 && r.activeStreamGen == gen
}

func (r *tickWsRuntime) hasActiveStream() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.activeStreamGen > 0
}

func buildTopicKey(topic types.WsTickTopic) string {
	topic = normalizeWsTickTopic(topic)
	return topic.Topic + "|" + topic.CategoryCode + "|" + topic.Symbol + "|" + topic.Market + "|" + topic.Interval
}

func normalizeWsTickTopic(topic types.WsTickTopic) types.WsTickTopic {
	topic.Topic = strings.ToLower(strings.TrimSpace(topic.Topic))
	topic.CategoryCode = strings.ToLower(strings.TrimSpace(topic.CategoryCode))
	topic.Symbol = strings.ToUpper(strings.TrimSpace(topic.Symbol))
	topic.Market = strings.ToUpper(strings.TrimSpace(topic.Market))
	topic.Interval = strings.ToLower(strings.TrimSpace(topic.Interval))
	if topic.Topic != "kline" {
		topic.Interval = ""
	}
	return topic
}

func isValidWsTickTopic(topic types.WsTickTopic) bool {
	if topic.Topic == "" || topic.CategoryCode == "" || topic.Symbol == "" {
		return false
	}
	if topic.Topic == "kline" && topic.Interval == "" {
		return false
	}
	return true
}
