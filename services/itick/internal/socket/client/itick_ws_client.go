package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	cutils "wklive/common/utils"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/internal/socket/server"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultWriteWait      = 10 * time.Second
	defaultPongWait       = 70 * time.Second
	defaultPingPeriod     = 30 * time.Second
	defaultReconnectDelay = 5 * time.Second
	defaultReconcileGap   = 5 * time.Second

	defaultLeaderTTL      = 15 * time.Second
	defaultLeaderRenewGap = 5 * time.Second
)

type ItickWsClient struct {
	url          string
	token        string
	categoryCode string

	dialer *websocket.Dialer

	mu   sync.RWMutex
	conn *websocket.Conn

	upstreamMu   sync.Mutex
	upstreamSubs map[string]struct{}

	bus      *ClusterBus
	registry *SubscriptionRegistry
	locker   *RedisLeaderLock
	hub      *server.Hub

	leader int32
	closed int32
}

func NewItickWsClient(
	url, token, categoryCode string,
	bus *ClusterBus,
	registry *SubscriptionRegistry,
	locker *RedisLeaderLock,
	hub *server.Hub,
) *ItickWsClient {
	return &ItickWsClient{
		url:          url,
		token:        token,
		categoryCode: categoryCode,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
		upstreamSubs: make(map[string]struct{}),
		bus:          bus,
		registry:     registry,
		locker:       locker,
		hub:          hub,
	}
}

func (c *ItickWsClient) Start(ctx context.Context) {
	go c.leaderLoop(ctx)
}

func (c *ItickWsClient) leaderLoop(ctx context.Context) {
	for {
		if ctx.Err() != nil || c.IsClosed() {
			return
		}

		lockCtx, cancel := context.WithCancel(ctx)

		token, err := c.locker.Acquire(lockCtx, defaultLeaderTTL)
		if err != nil {
			cancel()

			if errors.Is(err, ErrLockNotObtained) {
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					continue
				}
			}

			logx.Errorf("acquire itick leader lock failed, category=%s err=%v", c.categoryCode, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		atomic.StoreInt32(&c.leader, 1)
		logx.Infof("itick ws become leader, category=%s, url=%s", c.categoryCode, c.url)

		lostCh := make(chan struct{}, 1)

		go c.renewLoop(lockCtx, token, lostCh)

		if err := c.runAsLeader(lockCtx); err != nil && !errors.Is(err, context.Canceled) {
			logx.Errorf("itick ws leader session stopped, category=%s err=%v", c.categoryCode, err)
		}

		cancel()
		c.closeConn()
		atomic.StoreInt32(&c.leader, 0)

		select {
		case <-lostCh:
		default:
		}

		if err := c.locker.Release(context.Background(), token); err != nil {
			logx.Errorf("release itick leader lock failed, category=%s err=%v", c.categoryCode, err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *ItickWsClient) renewLoop(ctx context.Context, token string, lostCh chan<- struct{}) {
	ticker := time.NewTicker(defaultLeaderRenewGap)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ok, err := c.locker.Refresh(ctx, token, defaultLeaderTTL)
			if err != nil {
				logx.Errorf("refresh leader lock failed, category=%s err=%v", c.categoryCode, err)
				select {
				case lostCh <- struct{}{}:
				default:
				}
				return
			}
			if !ok {
				logx.Errorf("leader lock lost, category=%s", c.categoryCode)
				select {
				case lostCh <- struct{}{}:
				default:
				}
				return
			}
		}
	}
}

func (c *ItickWsClient) runAsLeader(ctx context.Context) error {
	for {
		if ctx.Err() != nil || c.IsClosed() {
			return ctx.Err()
		}

		if err := c.connect(); err != nil {
			logx.Errorf("itick ws connect failed: %v %s", err, c.url)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(defaultReconnectDelay):
				continue
			}
		}

		sessionCtx, stopSession := context.WithCancel(ctx)

		if err := c.restoreSubscriptions(sessionCtx); err != nil {
			logx.Errorf("itick ws restore subscriptions failed, category=%s err=%v", c.categoryCode, err)
		}

		go c.reconcileSubscriptionsLoop(sessionCtx)

		if err := c.readLoop(sessionCtx); err != nil {
			if isNormalWsClose(err) {
				logx.Errorf("itick ws read loop closed normally, category=%s err=%v", c.categoryCode, err)
			} else {
				logx.Errorf("itick ws read loop stopped, category=%s err=%v", c.categoryCode, err)
			}
		}

		stopSession()
		c.closeConn()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(defaultReconnectDelay):
		}
	}
}

func (c *ItickWsClient) reconcileSubscriptionsLoop(ctx context.Context) {
	ticker := time.NewTicker(defaultReconcileGap)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.restoreSubscriptions(ctx); err != nil {
				logx.Errorf("itick ws reconcile subscriptions failed, category=%s err=%v", c.categoryCode, err)
			}
		}
	}
}

func (c *ItickWsClient) connect() error {
	header := http.Header{}
	header.Set("token", c.token)

	conn, _, err := c.dialer.Dial(c.url, header)
	if err != nil {
		return err
	}

	_ = conn.SetReadDeadline(time.Now().Add(defaultPongWait))

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	c.resetUpstreamSubscriptions()

	go c.keepaliveLoop(conn)

	logx.Infof("itick ws connected: %s", c.url)
	return nil
}

func (c *ItickWsClient) keepaliveLoop(conn *websocket.Conn) {
	ticker := time.NewTicker(defaultPingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if c.IsClosed() {
			return
		}

		c.mu.RLock()
		current := c.conn
		c.mu.RUnlock()

		if current == nil || current != conn {
			return
		}

		ts := strconv.FormatInt(cutils.NowMillis(), 10)
		req := PingReq{
			Ac:     "ping",
			Params: ts,
		}

		_ = current.SetWriteDeadline(time.Now().Add(defaultWriteWait))
		if err := current.WriteJSON(req); err != nil {
			logx.Errorf("itick business ping failed: %v", err)
			return
		}
	}
}

func (c *ItickWsClient) readLoop(ctx context.Context) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("ws conn is nil")
	}

	for {
		if ctx.Err() != nil || c.IsClosed() {
			logx.Errorf("socket 链接关闭 ")
			return ctx.Err()
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			logx.Errorf("读取数据失败 %v", err)
			return err
		}

		_ = conn.SetReadDeadline(time.Now().Add(defaultPongWait))
		c.handleUpstreamMessage(ctx, data)
	}
}

func isNormalWsClose(err error) bool {
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
}

func (c *ItickWsClient) handleUpstreamMessage(ctx context.Context, data []byte) {
	var env UpstreamEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		logx.Errorf("itick ws unmarshal envelope failed: %v", err)
		return
	}

	if env.ResAc != "" {
		switch env.ResAc {
		case "auth", "subscribe", "unsubscribe":
			logx.Infof("itick control message: resAc=%s, code=%d, msg=%s", env.ResAc, env.Code, env.Msg)
		case "pong":
		default:
			logx.Infof("itick unknown control message: resAc=%s, code=%d, msg=%s", env.ResAc, env.Code, env.Msg)
		}
		return
	}

	if len(env.Data) == 0 {
		if env.Msg != "" {
			logx.Infof("itick message: code=%d, msg=%s", env.Code, env.Msg)
		}
		return
	}

	var d UpstreamData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		logx.Errorf("itick ws unmarshal data failed: %v", err)
		return
	}

	if d.S == "" || d.R == "" || d.Type == "" {
		logx.Errorf("数据异常")
		return
	}

	topic, interval := mapItickType(d.Type)
	if topic == "" {
		logx.Errorf("不支持的数据类型 %s", d.Type)
		return
	}

	msg := server.ClientMessage{
		Topic:        topic,
		CategoryCode: c.categoryCode,
		Symbol:       strings.ToUpper(strings.TrimSpace(d.S)),
		Market:       strings.ToUpper(strings.TrimSpace(d.R)),
		Interval:     interval,
	}

	switch topic {
	case server.TopicQuote:
		payload := QuotePayload{
			LastPrice: d.LD,
			Open:      d.O,
			High:      d.H,
			Low:       d.L,
			Volume:    d.V,
			Turnover:  d.TU,
			Ts:        d.T,
		}
		_ = c.bus.Publish(ctx, msg, &payload)

	case server.TopicTick:
		payload := TickPayload{
			LastPrice: d.LD,
			Volume:    d.V,
			Ts:        d.T,
		}
		_ = c.bus.Publish(ctx, msg, &payload)

	case server.TopicDepth:
		asks := make([]*DepthLevel, 0)
		bids := make([]*DepthLevel, 0)
		_ = json.Unmarshal(d.A, &asks)
		_ = json.Unmarshal(d.B, &bids)

		payload := DepthPayload{
			Asks: asks,
			Bids: bids,
		}
		_ = c.bus.Publish(ctx, msg, &payload)

	case server.TopicKline:
		payload := KlinePayload{
			Interval: interval,
			Open:     d.O,
			High:     d.H,
			Low:      d.L,
			Close:    d.C,
			Volume:   d.V,
			Turnover: d.TU,
			Ts:       d.T,
		}
		_ = c.bus.Publish(ctx, msg, &payload)
	}
}

func (c *ItickWsClient) HandleSubscriptionChange(change SubscriptionChange) {
	if !c.IsLeader() {
		return
	}
	if strings.ToLower(strings.TrimSpace(change.CategoryCode)) != c.categoryCode {
		return
	}

	msg := change.ToClientMessage()

	switch change.Action {
	case SubscriptionAdd:
		if err := c.subscribeByClientMessage(msg); err != nil {
			logx.Errorf("leader subscribe upstream failed, category=%s topic=%s err=%v", c.categoryCode, server.BuildTopicKey(msg), err)
		}
	case SubscriptionRemove:
		if err := c.UnsubscribeByClientMessage(msg); err != nil {
			logx.Errorf("leader unsubscribe upstream failed, category=%s topic=%s err=%v", c.categoryCode, server.BuildTopicKey(msg), err)
		}
	}
}

func (c *ItickWsClient) restoreSubscriptions(ctx context.Context) error {
	registryList, err := c.registry.ListActive(ctx, c.categoryCode)
	if err != nil {
		return err
	}

	merged := make(map[string]server.ClientMessage, len(registryList))
	for _, msg := range registryList {
		merged[server.BuildTopicKey(msg)] = msg
	}

	if c.hub != nil {
		for _, msg := range c.hub.SnapshotSubscriptions(c.categoryCode) {
			key := server.BuildTopicKey(msg)
			merged[key] = msg
			if err := c.registry.EnsureLease(ctx, msg); err != nil {
				logx.Errorf("restore local subscription lease failed, category=%s topic=%s err=%v", c.categoryCode, key, err)
			}
		}
	}

	for key, msg := range merged {
		if err := c.subscribeByClientMessage(msg); err != nil {
			logx.Errorf("restore subscribe failed, category=%s topic=%s err=%v", c.categoryCode, key, err)
		}
	}

	for key, msg := range c.snapshotUpstreamSubscriptions() {
		if _, ok := merged[key]; ok {
			continue
		}
		if err := c.UnsubscribeByClientMessage(msg); err != nil {
			logx.Errorf("restore unsubscribe stale topic failed, category=%s topic=%s err=%v", c.categoryCode, key, err)
		}
	}

	return nil
}

func (c *ItickWsClient) subscribeByClientMessage(msg server.ClientMessage) error {
	key := server.BuildTopicKey(msg)
	if c.hasUpstreamSubscription(key) {
		return nil
	}

	params, types, err := c.buildItickSubscribe(msg)
	if err != nil {
		return err
	}
	if err := c.subscribe(params, types); err != nil {
		return err
	}

	c.markUpstreamSubscription(key)
	return nil
}

func (c *ItickWsClient) UnsubscribeByClientMessage(msg server.ClientMessage) error {
	key := server.BuildTopicKey(msg)
	if !c.hasUpstreamSubscription(key) {
		return nil
	}

	params, types, err := c.buildItickSubscribe(msg)
	if err != nil {
		return err
	}
	if err := c.unsubscribe(params, types); err != nil {
		return err
	}

	c.unmarkUpstreamSubscription(key)
	return nil
}

func (c *ItickWsClient) buildItickSubscribe(msg server.ClientMessage) (params string, types string, err error) {
	if msg.Symbol == "" || msg.Market == "" {
		return "", "", errors.New("symbol or market is empty")
	}

	params = buildSymbolRegion(msg.Symbol, msg.Market)

	switch msg.Topic {
	case server.TopicQuote:
		types = "quote"
	case server.TopicDepth:
		types = "depth"
	case server.TopicTick:
		types = "tick"
	case server.TopicKline:
		types, err = intervalToItickKlineType(msg.Interval)
		if err != nil {
			return "", "", err
		}
	default:
		return "", "", fmt.Errorf("unsupported topic: %s", msg.Topic)
	}

	return params, types, nil
}

func (c *ItickWsClient) subscribe(params, types string) error {
	req := SubscribeReq{
		Ac:     "subscribe",
		Params: params,
		Types:  types,
	}

	if err := c.writeJSON(req); err != nil {
		return err
	}

	logx.Infof("itick subscribe success, category=%s params=%s, types=%s", c.categoryCode, params, types)
	return nil
}

func (c *ItickWsClient) unsubscribe(params, types string) error {
	req := UnsubscribeReq{
		Ac:     "unsubscribe",
		Params: params,
		Types:  types,
	}

	if err := c.writeJSON(req); err != nil {
		return err
	}

	logx.Infof("itick unsubscribe success, category=%s params=%s, types=%s", c.categoryCode, params, types)
	return nil
}

func (c *ItickWsClient) writeJSON(v any) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("ws conn is nil")
	}

	_ = conn.SetWriteDeadline(time.Now().Add(defaultWriteWait))
	return conn.WriteJSON(v)
}

func (c *ItickWsClient) Close() error {
	atomic.StoreInt32(&c.closed, 1)
	c.closeConn()
	return nil
}

func (c *ItickWsClient) IsClosed() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

func (c *ItickWsClient) IsLeader() bool {
	return atomic.LoadInt32(&c.leader) == 1
}

func (c *ItickWsClient) closeConn() {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()
	c.resetUpstreamSubscriptions()

	if conn != nil {
		_ = conn.Close()
	}
}

func (c *ItickWsClient) hasUpstreamSubscription(key string) bool {
	c.upstreamMu.Lock()
	defer c.upstreamMu.Unlock()

	_, ok := c.upstreamSubs[key]
	return ok
}

func (c *ItickWsClient) markUpstreamSubscription(key string) {
	c.upstreamMu.Lock()
	defer c.upstreamMu.Unlock()

	c.upstreamSubs[key] = struct{}{}
}

func (c *ItickWsClient) unmarkUpstreamSubscription(key string) {
	c.upstreamMu.Lock()
	defer c.upstreamMu.Unlock()

	delete(c.upstreamSubs, key)
}

func (c *ItickWsClient) resetUpstreamSubscriptions() {
	c.upstreamMu.Lock()
	defer c.upstreamMu.Unlock()

	c.upstreamSubs = make(map[string]struct{})
}

func (c *ItickWsClient) snapshotUpstreamSubscriptions() map[string]server.ClientMessage {
	c.upstreamMu.Lock()
	defer c.upstreamMu.Unlock()

	out := make(map[string]server.ClientMessage, len(c.upstreamSubs))
	for key := range c.upstreamSubs {
		out[key] = server.ParseTopicKey(key)
	}
	return out
}

func buildSymbolRegion(symbol, market string) string {
	return strings.ToUpper(strings.TrimSpace(symbol)) + "$" + strings.ToUpper(strings.TrimSpace(market))
}

func intervalToItickKlineType(interval string) (string, error) {
	return utils.IntervalToStream(interval)
}

func mapItickType(t string) (server.Topic, string) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "quote":
		return server.TopicQuote, ""
	case "depth":
		return server.TopicDepth, ""
	case "tick":
		return server.TopicTick, ""
	default:
		if interval, ok := utils.StreamToInterval(t); ok {
			return server.TopicKline, interval
		}
		return "", ""
	}
}
