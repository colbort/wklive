package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
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
	defaultAuthWait       = 10 * time.Second
	defaultPongWait       = 70 * time.Second
	defaultPingPeriod     = 30 * time.Second
	defaultReconnectDelay = 5 * time.Second
	defaultReconcileGap   = 5 * time.Second
	defaultSubscribeDelay = 50 * time.Millisecond

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

	writeMu sync.Mutex

	subMu          sync.Mutex
	desiredSubs    map[string]server.ClientMessage
	upstreamGroups map[string]string

	syncSubMu    sync.Mutex
	syncSubTimer *time.Timer

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
		desiredSubs:    make(map[string]server.ClientMessage),
		upstreamGroups: make(map[string]string),
		bus:            bus,
		registry:       registry,
		locker:         locker,
		hub:            hub,
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

		if err := c.waitAuthenticated(sessionCtx); err != nil {
			stopSession()
			c.closeConn()
			logx.Errorf("itick ws auth failed, category=%s err=%v", c.categoryCode, err)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(defaultReconnectDelay):
				continue
			}
		}

		if err := c.restoreSubscriptions(sessionCtx); err != nil {
			logx.Errorf("itick ws restore subscriptions failed, category=%s err=%v", c.categoryCode, err)
		}

		go c.reconcileSubscriptionsLoop(sessionCtx)

		if err := c.readLoop(sessionCtx); err != nil {
			if isNormalWsClose(err) {
				logx.Infof("itick ws read loop closed normally, category=%s err=%v", c.categoryCode, err)
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
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(defaultPongWait))
	})

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	c.resetUpstreamGroups()

	go c.keepaliveLoop(conn)

	logx.Errorf("itick ws connected: %s", c.url)
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

		if err := c.writePing(current); err != nil {
			logx.Errorf("itick ping failed: %v", err)
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
			logx.Infof("socket 链接关闭")
			return ctx.Err()
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			logx.Errorf("读取数据失败 category=%s %v", c.categoryCode, err)
			return err
		}

		_ = conn.SetReadDeadline(time.Now().Add(defaultPongWait))
		c.handleUpstreamMessage(ctx, data)
	}
}

func isNormalWsClose(err error) bool {
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
}

func (c *ItickWsClient) waitAuthenticated(ctx context.Context) error {
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
		if ctx.Err() != nil || c.IsClosed() {
			return ctx.Err()
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var env UpstreamEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("unmarshal auth message failed: %w", err)
		}

		if env.ResAc == "" {
			if env.Code == 1 && env.Msg != "" {
				logx.Infof("itick ws connected ack, category=%s msg=%s", c.categoryCode, env.Msg)
				continue
			}
			if len(env.Data) > 0 {
				c.handleUpstreamEnvelope(ctx, env)
				continue
			}
			if env.Msg != "" {
				logx.Infof("itick ws pre-auth message, category=%s code=%d msg=%s", c.categoryCode, env.Code, env.Msg)
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

		logx.Infof("itick ws authenticated, category=%s msg=%s", c.categoryCode, env.Msg)
		return nil
	}
}

func (c *ItickWsClient) handleUpstreamMessage(ctx context.Context, data []byte) {
	var env UpstreamEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		logx.Errorf("itick ws unmarshal envelope failed: %v", err)
		return
	}

	c.handleUpstreamEnvelope(ctx, env)
}

func (c *ItickWsClient) handleUpstreamEnvelope(ctx context.Context, env UpstreamEnvelope) {
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

func (c *ItickWsClient) HandleSubscriptionChanges(changes []SubscriptionChange) {
	if !c.IsLeader() {
		return
	}

	changed := false
	for _, change := range changes {
		if strings.ToLower(strings.TrimSpace(change.CategoryCode)) != c.categoryCode {
			continue
		}

		msg := change.ToClientMessage()
		key := server.BuildTopicKey(msg)

		switch change.Action {
		case SubscriptionAdd:
			if _, _, err := c.buildItickSubscribe(msg); err != nil {
				logx.Errorf("leader build subscribe failed, category=%s topic=%s err=%v", c.categoryCode, key, err)
				continue
			}
			if c.addDesiredSubscription(key, msg) {
				changed = true
			}
		case SubscriptionRemove:
			if c.removeDesiredSubscription(key) {
				changed = true
			}
		}
	}

	if changed {
		c.queueSubscriptionSync()
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

	return c.replaceDesiredSubscriptions(merged)
}

func (c *ItickWsClient) subscribeByClientMessages(items map[string]server.ClientMessage) error {
	return c.replaceDesiredSubscriptions(items)
}

func (c *ItickWsClient) replaceDesiredSubscriptions(items map[string]server.ClientMessage) error {
	next := make(map[string]server.ClientMessage, len(items))
	for key, msg := range items {
		if _, _, err := c.buildItickSubscribe(msg); err != nil {
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

	if !needSync {
		return nil
	}

	return c.syncDesiredSubscriptions()
}

func (c *ItickWsClient) addDesiredSubscription(key string, msg server.ClientMessage) bool {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	if old, ok := c.desiredSubs[key]; ok && sameClientMessage(old, msg) {
		return false
	}

	c.desiredSubs[key] = msg
	return true
}

func (c *ItickWsClient) removeDesiredSubscription(key string) bool {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	if _, ok := c.desiredSubs[key]; !ok {
		return false
	}

	delete(c.desiredSubs, key)
	return true
}

func (c *ItickWsClient) queueSubscriptionSync() {
	c.syncSubMu.Lock()
	defer c.syncSubMu.Unlock()

	if c.syncSubTimer != nil {
		c.syncSubTimer.Reset(defaultSubscribeDelay)
		return
	}

	c.syncSubTimer = time.AfterFunc(defaultSubscribeDelay, c.flushSubscriptionSync)
}

func (c *ItickWsClient) flushSubscriptionSync() {
	c.syncSubMu.Lock()
	c.syncSubTimer = nil
	c.syncSubMu.Unlock()

	if !c.IsLeader() || c.IsClosed() {
		return
	}

	if err := c.syncDesiredSubscriptions(); err != nil {
		logx.Errorf("sync desired subscriptions failed, category=%s err=%v", c.categoryCode, err)
	}
}

func (c *ItickWsClient) syncDesiredSubscriptions() error {
	c.subMu.Lock()
	desired := make(map[string]server.ClientMessage, len(c.desiredSubs))
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

func (c *ItickWsClient) buildSubscriptionGroups(items map[string]server.ClientMessage) (map[string]string, error) {
	groupSets := make(map[string]map[string]struct{})
	for key, msg := range items {
		params, types, err := c.buildItickSubscribe(msg)
		if err != nil {
			return nil, fmt.Errorf("build subscribe failed, topic=%s err=%w", key, err)
		}
		if groupSets[types] == nil {
			groupSets[types] = make(map[string]struct{})
		}
		groupSets[types][params] = struct{}{}
	}

	groups := make(map[string]string, len(groupSets))
	for types, set := range groupSets {
		params := make([]string, 0, len(set))
		for item := range set {
			params = append(params, item)
		}
		sort.Strings(params)
		groups[types] = strings.Join(params, ",")
	}

	return groups, nil
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

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	_ = conn.SetWriteDeadline(time.Now().Add(defaultWriteWait))
	return conn.WriteJSON(v)
}

func (c *ItickWsClient) writePing(conn *websocket.Conn) error {
	ts := strconv.FormatInt(cutils.NowMillis(), 10)
	req := PingReq{
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
	c.resetPendingSubscriptionSync()
	c.resetUpstreamGroups()

	if conn != nil {
		_ = conn.Close()
	}
}

func (c *ItickWsClient) resetPendingSubscriptionSync() {
	c.syncSubMu.Lock()
	defer c.syncSubMu.Unlock()

	if c.syncSubTimer != nil {
		c.syncSubTimer.Stop()
		c.syncSubTimer = nil
	}
}

func (c *ItickWsClient) resetUpstreamGroups() {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	c.upstreamGroups = make(map[string]string)
}

func sameDesiredSubscriptions(left, right map[string]server.ClientMessage) bool {
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

func sameClientMessage(left, right server.ClientMessage) bool {
	return left.Topic == right.Topic &&
		left.CategoryCode == right.CategoryCode &&
		left.Symbol == right.Symbol &&
		left.Market == right.Market &&
		left.Interval == right.Interval
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
