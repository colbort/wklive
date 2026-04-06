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
	"time"

	"wklive/services/itick/internal/socket/server"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultWriteWait      = 10 * time.Second
	defaultPongWait       = 70 * time.Second
	defaultPingPeriod     = 30 * time.Second
	defaultReconnectDelay = 5 * time.Second
)

type ItickWsClient struct {
	url          string
	token        string
	categoryCode string // 固定 crypto

	dialer *websocket.Dialer

	mu   sync.RWMutex
	conn *websocket.Conn

	closed bool

	hub *server.Hub

	// 断线重连后恢复订阅
	subMu       sync.RWMutex
	subscribers map[string]SubscribeReq
}

func NewItickWsClient(url, token string, categoryCode string, hub *server.Hub) *ItickWsClient {
	return &ItickWsClient{
		url:          url,
		token:        token,
		categoryCode: categoryCode,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
		hub:         hub,
		subscribers: make(map[string]SubscribeReq),
	}
}

func (c *ItickWsClient) Start(ctx context.Context) {
	go c.loop(ctx)
}

func (c *ItickWsClient) loop(ctx context.Context) {
	for {
		if ctx.Err() != nil || c.IsClosed() {
			return
		}

		if err := c.connect(); err != nil {
			logx.Errorf("itick ws connect failed: %v  %s", err, c.url)
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultReconnectDelay):
				continue
			}
		}

		if err := c.restoreSubscriptions(); err != nil {
			logx.Errorf("itick ws restore subscriptions failed: %v", err)
		}

		if err := c.readLoop(ctx); err != nil {
			logx.Errorf("itick ws read loop stopped: %v", err)
		}

		c.closeConn()

		select {
		case <-ctx.Done():
			return
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

		ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
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
		c.handleUpstreamMessage(data)
	}
}

func (c *ItickWsClient) handleUpstreamMessage(data []byte) {
	var env UpstreamEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		logx.Errorf("itick ws unmarshal envelope failed: %v", err)
		return
	}

	// 控制消息
	if env.ResAc != "" {
		switch env.ResAc {
		case "auth", "subscribe", "unsubscribe":
			logx.Infof("itick control message: resAc=%s, code=%d, msg=%s", env.ResAc, env.Code, env.Msg)
		case "pong":
			// 心跳响应
		default:
			logx.Infof("itick unknown control message: resAc=%s, code=%d, msg=%s", env.ResAc, env.Code, env.Msg)
		}
		return
	}

	if len(env.Data) == 0 {
		// Connected Successfully 之类
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
		c.hub.Broadcast(msg, &payload)

	case server.TopicTick:
		payload := TickPayload{
			LastPrice: d.LD,
			Volume:    d.V,
			Ts:        d.T,
		}
		c.hub.Broadcast(msg, &payload)

	case server.TopicDepth:
		asks := make([]*DepthLevel, 0)
		bids := make([]*DepthLevel, 0)
		_ = json.Unmarshal(d.A, &asks)
		_ = json.Unmarshal(d.B, &bids)

		payload := DepthPayload{
			Asks: asks,
			Bids: bids,
		}

		c.hub.Broadcast(msg, &payload)

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
		c.hub.Broadcast(msg, &payload)
	}
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
	return c.Unsubscribe(params, types)
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

	key := buildSubKey(params, types)

	if err := c.writeJSON(req); err != nil {
		return err
	}

	c.subMu.Lock()
	c.subscribers[key] = req
	c.subMu.Unlock()

	logx.Infof("itick subscribe success, params=%s, types=%s", params, types)
	return nil
}

func (c *ItickWsClient) Unsubscribe(params, types string) error {
	req := UnsubscribeReq{
		Ac:     "unsubscribe",
		Params: params,
		Types:  types,
	}

	if err := c.writeJSON(req); err != nil {
		return err
	}

	key := buildSubKey(params, types)

	c.subMu.Lock()
	delete(c.subscribers, key)
	c.subMu.Unlock()

	logx.Infof("itick unsubscribe success, params=%s, types=%s", params, types)
	return nil
}

func (c *ItickWsClient) restoreSubscriptions() error {
	c.subMu.RLock()
	list := make([]SubscribeReq, 0, len(c.subscribers))
	for _, sub := range c.subscribers {
		list = append(list, sub)
	}
	c.subMu.RUnlock()

	for _, sub := range list {
		if err := c.writeJSON(sub); err != nil {
			return err
		}
	}

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
	c.mu.Lock()
	c.closed = true
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn != nil {
		return conn.Close()
	}
	return nil
}

func (c *ItickWsClient) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
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

func buildSubKey(params, types string) string {
	return params + "|" + types
}

func buildSymbolRegion(symbol, market string) string {
	return strings.ToUpper(strings.TrimSpace(symbol)) + "$" + strings.ToUpper(strings.TrimSpace(market))
}

func intervalToItickKlineType(interval string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "1m":
		return "kline@1", nil
	case "5m":
		return "kline@2", nil
	case "15m":
		return "kline@3", nil
	case "30m":
		return "kline@4", nil
	case "1h", "60m":
		return "kline@5", nil
	case "1d":
		return "kline@8", nil
	case "1w":
		return "kline@9", nil
	case "1mo":
		return "kline@10", nil
	default:
		return "", fmt.Errorf("unsupported interval: %s", interval)
	}
}

func mapItickType(t string) (server.Topic, string) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "quote":
		return server.TopicQuote, ""
	case "depth":
		return server.TopicDepth, ""
	case "tick":
		return server.TopicTick, ""
	case "kline@1":
		return server.TopicKline, "1m"
	case "kline@2":
		return server.TopicKline, "5m"
	case "kline@3":
		return server.TopicKline, "15m"
	case "kline@4":
		return server.TopicKline, "30m"
	case "kline@5":
		return server.TopicKline, "1h"
	case "kline@8":
		return server.TopicKline, "1d"
	case "kline@9":
		return server.TopicKline, "1w"
	case "kline@10":
		return server.TopicKline, "1mo"
	default:
		return "", ""
	}
}
