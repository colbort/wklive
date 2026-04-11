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

	bus      *ClusterBus
	registry *SubscriptionRegistry
	locker   *RedisLeaderLock

	leader int32
	closed int32
}

func NewItickWsClient(
	url, token, categoryCode string,
	bus *ClusterBus,
	registry *SubscriptionRegistry,
	locker *RedisLeaderLock,
) *ItickWsClient {
	return &ItickWsClient{
		url:          url,
		token:        token,
		categoryCode: categoryCode,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
		bus:      bus,
		registry: registry,
		locker:   locker,
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

		if err := c.restoreSubscriptions(ctx); err != nil {
			logx.Errorf("itick ws restore subscriptions failed, category=%s err=%v", c.categoryCode, err)
		}

		if err := c.readLoop(ctx); err != nil {
			logx.Errorf("itick ws read loop stopped, category=%s err=%v", c.categoryCode, err)
		}

		c.closeConn()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(defaultReconnectDelay):
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

		ts := strconv.FormatInt(utils.NowMillis(), 10)
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
			return ctx.Err()
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		_ = conn.SetReadDeadline(time.Now().Add(defaultPongWait))
		c.handleUpstreamMessage(ctx, data)
	}
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
		return
	}

	topic, interval := mapItickType(d.Type)
	if topic == "" {
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
		if err := c.SubscribeByClientMessage(msg); err != nil {
			logx.Errorf("leader subscribe upstream failed, category=%s topic=%s err=%v", c.categoryCode, server.BuildTopicKey(msg), err)
		}
	case SubscriptionRemove:
		if err := c.UnsubscribeByClientMessage(msg); err != nil {
			logx.Errorf("leader unsubscribe upstream failed, category=%s topic=%s err=%v", c.categoryCode, server.BuildTopicKey(msg), err)
		}
	}
}

func (c *ItickWsClient) restoreSubscriptions(ctx context.Context) error {
	list, err := c.registry.ListActive(ctx, c.categoryCode)
	if err != nil {
		return err
	}

	for _, msg := range list {
		if err := c.SubscribeByClientMessage(msg); err != nil {
			logx.Errorf("restore subscribe failed, category=%s topic=%s err=%v", c.categoryCode, server.BuildTopicKey(msg), err)
		}
	}

	return nil
}

func (c *ItickWsClient) SubscribeByClientMessage(msg server.ClientMessage) error {
	params, types, err := c.buildItickSubscribe(msg)
	if err != nil {
		return err
	}
	return c.subscribe(params, types)
}

func (c *ItickWsClient) UnsubscribeByClientMessage(msg server.ClientMessage) error {
	params, types, err := c.buildItickSubscribe(msg)
	if err != nil {
		return err
	}
	return c.unsubscribe(params, types)
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

	if conn != nil {
		_ = conn.Close()
	}
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
