package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	cutils "wklive/common/utils"
	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/types"
	"wklive/services/market/internal/pkg/utils"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultWriteWait      = 10 * time.Second
	defaultAuthWait       = 10 * time.Second
	defaultPongWait       = 70 * time.Second
	defaultPingPeriod     = 30 * time.Second
	defaultReconnectDelay = 5 * time.Second

	defaultLeaderTTL      = 15 * time.Second
	defaultLeaderRenewGap = 5 * time.Second
)

type wsHandshakeError struct {
	err        error
	statusCode int
	body       string
	retryAfter time.Duration
}

func (e *wsHandshakeError) Error() string {
	return fmt.Sprintf("%v: http_status=%d body=%q", e.err, e.statusCode, e.body)
}

func (e *wsHandshakeError) Unwrap() error { return e.err }

type ITickWsClient struct {
	url          string
	token        string
	categoryCode string

	dialer *websocket.Dialer

	mu   sync.RWMutex
	conn *websocket.Conn

	writeMu sync.Mutex

	subMu          sync.Mutex
	desiredSubs    map[string]types.ClientMessage
	upstreamGroups map[string]string

	marketCache    *cache.MarketDataCache
	locker         *RedisLeaderLock
	connectLimiter *RedisConnectLimiter

	leader        int32
	started       int32
	authenticated int32
	sessions      int64
	onReconnect   func(string)
}

func NewMarketWsClient(
	url, token, categoryCode string,
	marketCache *cache.MarketDataCache,
	locker *RedisLeaderLock,
	connectLimiter *RedisConnectLimiter,
) *ITickWsClient {
	return &ITickWsClient{
		url:          url,
		token:        token,
		categoryCode: categoryCode,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
		desiredSubs:    make(map[string]types.ClientMessage),
		upstreamGroups: make(map[string]string),
		marketCache:    marketCache,
		locker:         locker,
		connectLimiter: connectLimiter,
	}
}

func (c *ITickWsClient) Start(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		return
	}
	go c.leaderLoop(ctx)
}

func (c *ITickWsClient) HasDesiredSubscriptions() bool {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	return len(c.desiredSubs) > 0
}

func (c *ITickWsClient) leaderLoop(ctx context.Context) {
	for {
		if ctx.Err() != nil {
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

			logx.Errorf("acquire market leader lock failed, category=%s err=%v", c.categoryCode, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		atomic.StoreInt32(&c.leader, 1)
		logx.Infof("market ws become leader, category=%s, url=%s", c.categoryCode, c.url)

		lostCh := make(chan struct{}, 1)

		go c.renewLoop(lockCtx, token, lostCh, cancel)

		if err := c.runAsLeader(lockCtx); err != nil && !errors.Is(err, context.Canceled) {
			logx.Errorf("market ws leader session stopped, category=%s err=%v", c.categoryCode, err)
		}

		cancel()
		c.closeConn()
		atomic.StoreInt32(&c.leader, 0)

		select {
		case <-lostCh:
		default:
		}

		if err := c.locker.Release(context.Background(), token); err != nil {
			logx.Errorf("release market leader lock failed, category=%s err=%v", c.categoryCode, err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *ITickWsClient) renewLoop(ctx context.Context, token string, lostCh chan<- struct{}, cancel context.CancelFunc) {
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
				c.handleLeaderLost(lostCh, cancel)
				return
			}
			if !ok {
				logx.Errorf("leader lock lost, category=%s", c.categoryCode)
				c.handleLeaderLost(lostCh, cancel)
				return
			}
		}
	}
}

func (c *ITickWsClient) handleLeaderLost(lostCh chan<- struct{}, cancel context.CancelFunc) {
	select {
	case lostCh <- struct{}{}:
	default:
	}
	cancel()
	c.closeConn()
}

func (c *ITickWsClient) runAsLeader(ctx context.Context) error {
	reconnectDelay := defaultReconnectDelay
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if c.connectLimiter != nil {
			if err := c.connectLimiter.Wait(ctx); err != nil {
				return err
			}
		}
		if err := c.connect(); err != nil {
			logx.Errorf("market ws connect failed: %v %s", err, c.url)
			var handshakeErr *wsHandshakeError
			if errors.As(err, &handshakeErr) && handshakeErr.statusCode == http.StatusTooManyRequests && c.connectLimiter != nil {
				if penalizeErr := c.connectLimiter.Penalize(ctx, handshakeErr.retryAfter); penalizeErr != nil {
					logx.Errorf("set market ws global cool down failed: %v", penalizeErr)
				}
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(reconnectDelay + time.Duration(rand.Int63n(int64(time.Second)))):
				reconnectDelay = min(reconnectDelay*2, time.Minute)
				continue
			}
		}
		reconnectDelay = defaultReconnectDelay

		sessionCtx, stopSession := context.WithCancel(ctx)

		if err := c.waitAuthenticated(sessionCtx); err != nil {
			stopSession()
			c.closeConn()
			logx.Errorf("market ws auth failed, category=%s err=%v", c.categoryCode, err)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(defaultReconnectDelay):
				continue
			}
		}
		atomic.StoreInt32(&c.authenticated, 1)

		if err := c.restoreSubscriptions(sessionCtx); err != nil {
			logx.Errorf("market ws restore subscriptions failed, category=%s err=%v", c.categoryCode, err)
		}
		atomic.AddInt64(&c.sessions, 1)
		if c.onReconnect != nil {
			go c.onReconnect(c.categoryCode)
		}

		if err := c.readLoop(sessionCtx); err != nil {
			if isNormalWsClose(err) {
				logx.Infof("market ws read loop closed normally, category=%s err=%v", c.categoryCode, err)
			} else {
				logx.Errorf("market ws read loop stopped, category=%s err=%v", c.categoryCode, err)
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

func (c *ITickWsClient) SetReconnectHandler(handler func(string)) {
	c.onReconnect = handler
}

func (c *ITickWsClient) connect() error {
	header := http.Header{}
	header.Set("token", c.token)

	conn, resp, err := c.dialer.Dial(c.url, header)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			return &wsHandshakeError{
				err: err, statusCode: resp.StatusCode, body: strings.TrimSpace(string(body)),
				retryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
			}
		}
		return err
	}

	_ = conn.SetReadDeadline(time.Now().Add(defaultPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(defaultPongWait))
	})

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	c.resetUpstreamGroups()

	go c.keepaliveLoop(conn)

	logx.Infof("market ws connected: %s", c.url)
	return nil
}

func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if retryAt, err := http.ParseTime(value); err == nil {
		return max(time.Until(retryAt), 0)
	}
	return 30 * time.Second
}

func (c *ITickWsClient) keepaliveLoop(conn *websocket.Conn) {
	ticker := time.NewTicker(defaultPingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.RLock()
		current := c.conn
		c.mu.RUnlock()

		if current == nil || current != conn {
			return
		}

		if err := c.writePing(current); err != nil {
			logx.Errorf("market ping failed, category=%s err=%v", c.categoryCode, err)
			c.closeConn()
			return
		}
	}
}

func (c *ITickWsClient) readLoop(ctx context.Context) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("ws conn is nil")
	}

	for {
		if ctx.Err() != nil {
			logx.Infof("socket 链接关闭")
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

func isNormalWsClose(err error) bool {
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
}

func (c *ITickWsClient) waitAuthenticated(ctx context.Context) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("ws conn is nil")
	}

	authDeadline := time.Now().Add(defaultAuthWait)
	_ = conn.SetReadDeadline(authDeadline)
	defer func() {
		_ = conn.SetReadDeadline(time.Now().Add(defaultPongWait))
	}()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var env types.UpstreamEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("unmarshal auth message failed: %w", err)
		}

		if env.ResAc == "" {
			if env.Code == 1 && env.Msg != "" {
				logx.Infof("market ws connected ack, category=%s msg=%s", c.categoryCode, env.Msg)
				continue
			}
			if len(env.Data) > 0 {
				c.handleUpstreamEnvelope(ctx, env)
				continue
			}
			if env.Msg != "" {
				logx.Infof("market ws pre-auth message, category=%s code=%d msg=%s", c.categoryCode, env.Code, env.Msg)
			}
			continue
		}

		if env.ResAc != "auth" {
			c.handleUpstreamEnvelope(ctx, env)
			continue
		}

		if env.Code != 1 {
			return fmt.Errorf("auth rejected: code=%d msg=%s", env.Code, env.Msg)
		}

		logx.Infof("market ws authenticated, category=%s msg=%s", c.categoryCode, env.Msg)
		return nil
	}
}

func (c *ITickWsClient) handleUpstreamMessage(ctx context.Context, data []byte) {
	var env types.UpstreamEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		logx.Errorf("market ws unmarshal envelope failed: %v", err)
		return
	}

	c.handleUpstreamEnvelope(ctx, env)
}

func (c *ITickWsClient) handleUpstreamEnvelope(ctx context.Context, env types.UpstreamEnvelope) {
	if env.ResAc != "" {
		switch env.ResAc {
		case "auth", "subscribe", "unsubscribe":
			logx.Infof("market control message: resAc=%s, code=%d, msg=%s", env.ResAc, env.Code, env.Msg)
		case "pong":
		default:
			logx.Infof("market unknown control message: resAc=%s, code=%d, msg=%s", env.ResAc, env.Code, env.Msg)
		}
		return
	}

	if len(env.Data) == 0 {
		if env.Msg != "" {
			logx.Infof("market message: code=%d, msg=%s", env.Code, env.Msg)
		}
		return
	}

	var d types.UpstreamData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		logx.Errorf("market ws unmarshal data failed: %v", err)
		return
	}

	if d.S == "" || d.R == "" || d.Type == "" {
		logx.Errorf("数据异常")
		return
	}

	topic, interval := mapMarketType(d.Type)
	if topic == "" {
		logx.Errorf("不支持的数据类型 %s", d.Type)
		return
	}

	msg := types.ClientMessage{
		Topic:        topic,
		CategoryCode: c.categoryCode,
		Symbol:       strings.ToUpper(strings.TrimSpace(d.S)),
		Market:       strings.ToUpper(strings.TrimSpace(d.R)),
		Interval:     interval,
	}

	switch topic {
	case types.TopicQuote:
		payload := types.QuotePayload{
			LastPrice:     d.LD,
			LastPriceText: rawDecimalToken(env.Data, "ld"),
			Open:          d.O,
			High:          d.H,
			Low:           d.L,
			Volume:        d.V,
			Turnover:      d.TU,
			Ts:            d.T,
			Authority:     "market-ws",
		}
		if err := c.marketCache.Set(ctx, msg, &payload); err != nil {
			logx.Errorf("cache market quote failed, category=%s market=%s symbol=%s err=%v", msg.CategoryCode, msg.Market, msg.Symbol, err)
		}

	case types.TopicTick:
		payload := types.TickPayload{
			LastPrice: d.LD,
			Volume:    d.V,
			Turnover:  d.TU,
			Ts:        d.T,
		}
		if err := c.marketCache.Set(ctx, msg, &payload); err != nil {
			logx.Errorf("cache market tick failed, category=%s market=%s symbol=%s err=%v", msg.CategoryCode, msg.Market, msg.Symbol, err)
		}

	case types.TopicDepth:
		asks := make([]*types.DepthLevel, 0)
		bids := make([]*types.DepthLevel, 0)
		_ = json.Unmarshal(d.A, &asks)
		_ = json.Unmarshal(d.B, &bids)
		asks = appendSyntheticDepthLevels(asks, true, 6)
		bids = appendSyntheticDepthLevels(bids, false, 6)

		payload := types.DepthPayload{
			Asks: asks,
			Bids: bids,
		}
		if err := c.marketCache.Set(ctx, msg, &payload); err != nil {
			logx.Errorf("cache market depth failed, category=%s market=%s symbol=%s err=%v", msg.CategoryCode, msg.Market, msg.Symbol, err)
		}

	case types.TopicKline:
		payload := types.KlinePayload{
			Interval: interval,
			Open:     d.O,
			High:     d.H,
			Low:      d.L,
			Close:    d.C,
			Volume:   d.V,
			Turnover: d.TU,
			Ts:       d.T,
			Source:   "market_ws",
			Revision: time.Now().UnixMilli(),
		}
		if err := c.marketCache.Set(ctx, msg, &payload); err != nil {
			logx.Errorf("cache market kline failed, category=%s market=%s symbol=%s interval=%s err=%v", msg.CategoryCode, msg.Market, msg.Symbol, msg.Interval, err)
		}
	}
}

func rawDecimalToken(data json.RawMessage, field string) string {
	var values map[string]json.RawMessage
	if json.Unmarshal(data, &values) != nil {
		return ""
	}
	raw := strings.TrimSpace(string(values[field]))
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	return raw
}

func (c *ITickWsClient) restoreSubscriptions(_ context.Context) error {
	return c.syncDesiredSubscriptions()
}

func (c *ITickWsClient) replaceDesiredSubscriptions(items map[string]types.ClientMessage) error {
	next := make(map[string]types.ClientMessage, len(items))
	for key, msg := range items {
		if _, _, err := c.buildMarketSubscribe(msg); err != nil {
			logx.Errorf("build desired subscribe failed, category=%s topic=%s err=%v", c.categoryCode, key, err)
			continue
		}
		next[key] = msg
	}

	c.subMu.Lock()
	changed := !sameDesiredSubscriptions(c.desiredSubs, next)
	needSync := changed || (len(c.upstreamGroups) == 0 && len(next) > 0)
	if changed {
		c.desiredSubs = next
	}
	c.subMu.Unlock()

	if !needSync || !c.canSyncSubscriptions() {
		return nil
	}

	return c.syncDesiredSubscriptions()
}

func (c *ITickWsClient) syncDesiredSubscriptions() error {
	c.subMu.Lock()
	desired := make(map[string]types.ClientMessage, len(c.desiredSubs))
	for key, msg := range c.desiredSubs {
		desired[key] = msg
	}
	previous := make(map[string]string, len(c.upstreamGroups))
	for types, params := range c.upstreamGroups {
		previous[types] = params
	}
	c.subMu.Unlock()

	next, err := c.buildSubscriptionGroups(desired)
	if err != nil {
		return err
	}

	for types, oldParams := range previous {
		if newParams, ok := next[types]; ok && newParams == oldParams {
			continue
		}
		if oldParams == "" {
			continue
		}
		if err := c.unsubscribe(oldParams, types); err != nil {
			return err
		}
	}

	for types, params := range next {
		if params == "" || previous[types] == params {
			continue
		}
		if err := c.subscribe(params, types); err != nil {
			return err
		}
	}

	c.subMu.Lock()
	c.upstreamGroups = next
	c.subMu.Unlock()
	return nil
}

func appendSyntheticDepthLevels(levels []*types.DepthLevel, isAsk bool, count int) []*types.DepthLevel {
	if count <= 0 {
		return levels
	}

	valid := make([]*types.DepthLevel, 0, len(levels))
	for _, level := range levels {
		if level != nil && level.Price > 0 {
			valid = append(valid, level)
		}
	}
	if len(valid) == 0 {
		return levels
	}

	prices := make([]float64, 0, len(valid))
	totalVolume := 0.0
	var maxPosition int64
	for _, level := range valid {
		prices = append(prices, level.Price)
		if level.Volume > 0 {
			totalVolume += level.Volume
		}
		if level.Position > maxPosition {
			maxPosition = level.Position
		}
	}
	sort.Float64s(prices)

	step := depthPriceStep(prices)
	avgVolume := totalVolume / float64(len(valid))
	if avgVolume <= 0 {
		avgVolume = 1
	}

	basePrice := prices[0]
	direction := -1.0
	if isAsk {
		basePrice = prices[len(prices)-1]
		direction = 1
	}

	out := append([]*types.DepthLevel{}, levels...)
	for i := 1; i <= count; i++ {
		priceOffset := step * float64(i) * (1 + rand.Float64()*0.35)
		price := basePrice + direction*priceOffset
		if price <= 0 {
			continue
		}
		volume := avgVolume * (0.55 + rand.Float64()*0.9)
		out = append(out, &types.DepthLevel{
			Price:        roundDepthPrice(price),
			Volume:       roundDepthVolume(volume),
			Position:     maxPosition + int64(i),
			OriginVolume: roundDepthVolume(volume),
		})
	}

	return out
}

func depthPriceStep(prices []float64) float64 {
	if len(prices) < 2 {
		return math.Max(roundDepthPrice(prices[0]*0.0001), 0.01)
	}

	minStep := math.MaxFloat64
	for i := 1; i < len(prices); i++ {
		step := math.Abs(prices[i] - prices[i-1])
		if step > 0 && step < minStep {
			minStep = step
		}
	}
	if minStep == math.MaxFloat64 || minStep <= 0 {
		return math.Max(roundDepthPrice(prices[0]*0.0001), 0.01)
	}
	return minStep
}

func roundDepthPrice(value float64) float64 {
	return math.Round(value*100) / 100
}

func roundDepthVolume(value float64) float64 {
	return math.Round(value*1e5) / 1e5
}

func (c *ITickWsClient) buildSubscriptionGroups(items map[string]types.ClientMessage) (map[string]string, error) {
	groupSets := make(map[string]map[string]struct{})
	for key, msg := range items {
		params, types, err := c.buildMarketSubscribe(msg)
		if err != nil {
			return nil, fmt.Errorf("build subscribe failed, topic=%s err=%w", key, err)
		}
		if groupSets[types] == nil {
			groupSets[types] = make(map[string]struct{})
		}
		groupSets[types][params] = struct{}{}
	}

	typesByParams := make(map[string][]string)
	for streamType, set := range groupSets {
		params := make([]string, 0, len(set))
		for item := range set {
			params = append(params, item)
		}
		sort.Strings(params)
		paramsKey := strings.Join(params, ",")
		typesByParams[paramsKey] = append(typesByParams[paramsKey], streamType)
	}

	groups := make(map[string]string, len(typesByParams))
	for params, streamTypes := range typesByParams {
		sort.Strings(streamTypes)
		groups[strings.Join(streamTypes, ",")] = params
	}

	return groups, nil
}

func (c *ITickWsClient) buildMarketSubscribe(msg types.ClientMessage) (string, string, error) {
	if msg.Symbol == "" || msg.Market == "" {
		return "", "", errors.New("symbol or market is empty")
	}

	params := buildSymbolRegion(msg.Symbol, msg.Market)
	tys := ""

	switch msg.Topic {
	case types.TopicQuote:
		tys = "quote"
	case types.TopicDepth:
		tys = "depth"
	case types.TopicTick:
		tys = "tick"
	case types.TopicKline:
		temp, err := utils.IntervalToStream(msg.Interval)
		if err != nil {
			return "", "", err
		}
		tys = temp
	default:
		return "", "", fmt.Errorf("unsupported topic: %s", msg.Topic)
	}

	return params, tys, nil
}

func (c *ITickWsClient) subscribe(params string, tys string) error {
	req := types.SubscribeReq{
		Ac:     "subscribe",
		Params: params,
		Types:  tys,
	}

	if err := c.writeJSON(req); err != nil {
		logx.Errorf("market subscribe failed, category=%s params=%s types=%s err=%v", c.categoryCode, params, tys, err)
		return err
	}

	return nil
}

func (c *ITickWsClient) unsubscribe(params string, tys string) error {
	req := types.UnsubscribeReq{
		Ac:     "unsubscribe",
		Params: params,
		Types:  tys,
	}

	if err := c.writeJSON(req); err != nil {
		return err
	}

	logx.Infof("market unsubscribe success, category=%s params=%s, types=%s", c.categoryCode, params, tys)
	return nil
}

func (c *ITickWsClient) writeJSON(v any) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("ws conn is nil")
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	_ = conn.SetWriteDeadline(time.Now().Add(defaultWriteWait))
	return conn.WriteJSON(v)
}

func (c *ITickWsClient) writePing(conn *websocket.Conn) error {
	ts := strconv.FormatInt(cutils.NowMillis(), 10)
	req := types.PingReq{
		Ac:     "ping",
		Params: ts,
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	deadline := time.Now().Add(defaultWriteWait)
	if err := conn.WriteControl(websocket.PingMessage, []byte(ts), deadline); err != nil {
		return err
	}

	_ = conn.SetWriteDeadline(deadline)
	return conn.WriteJSON(req)
}

func (c *ITickWsClient) IsLeader() bool {
	return atomic.LoadInt32(&c.leader) == 1
}

func (c *ITickWsClient) canSyncSubscriptions() bool {
	if !c.IsLeader() || atomic.LoadInt32(&c.authenticated) != 1 {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}

func (c *ITickWsClient) closeConn() {
	atomic.StoreInt32(&c.authenticated, 0)
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()
	c.resetUpstreamGroups()

	if conn != nil {
		_ = conn.Close()
	}
}

func (c *ITickWsClient) resetUpstreamGroups() {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	c.upstreamGroups = make(map[string]string)
}

func sameDesiredSubscriptions(left map[string]types.ClientMessage, right map[string]types.ClientMessage) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftMsg := range left {
		rightMsg, ok := right[key]
		if !ok || !sameClientMessage(leftMsg, rightMsg) {
			return false
		}
	}
	return true
}

func sameClientMessage(left types.ClientMessage, right types.ClientMessage) bool {
	return left.Topic == right.Topic &&
		left.CategoryCode == right.CategoryCode &&
		left.Symbol == right.Symbol &&
		left.Market == right.Market &&
		left.Interval == right.Interval
}

func buildSymbolRegion(symbol string, market string) string {
	return strings.ToUpper(strings.TrimSpace(symbol)) + "$" + strings.ToUpper(strings.TrimSpace(market))
}

func mapMarketType(t string) (types.Topic, string) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "quote":
		return types.TopicQuote, ""
	case "depth":
		return types.TopicDepth, ""
	case "tick":
		return types.TopicTick, ""
	default:
		if interval, ok := utils.StreamToInterval(t); ok {
			return types.TopicKline, interval
		}
		return "", ""
	}
}
