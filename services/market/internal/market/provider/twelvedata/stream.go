package twelvedata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/provider"
	"wklive/services/market/internal/market/types"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	wsWriteWait       = 10 * time.Second
	wsReadWait        = 45 * time.Second
	wsHeartbeatPeriod = 10 * time.Second
	wsReconnectDelay  = 5 * time.Second
	leaderTTL         = 15 * time.Second
	leaderRenewPeriod = 5 * time.Second
	// A single event can subscribe multiple symbols. Batching keeps payloads
	// bounded while staying far below the 100 client-events/minute limit.
	wsSubscribeBatch = 500
)

type Stream struct {
	url          string
	apiKey       string
	categoryCode string
	marketCache  *cache.MarketDataCache
	locker       *redisLeaderLock
	dialer       *websocket.Dialer

	mu      sync.RWMutex
	conn    *websocket.Conn
	writeMu sync.Mutex

	subMu   sync.Mutex
	desired map[string]provider.Subscription
	sent    map[string]struct{}

	leader      int32
	started     int32
	connected   int32
	onReconnect func(string)
}

func newStream(rawURL, apiKey string, marketCache *cache.MarketDataCache, locker *redisLeaderLock) *Stream {
	return &Stream{
		url:          strings.TrimSpace(rawURL),
		apiKey:       strings.TrimSpace(apiKey),
		categoryCode: supportedCategory,
		marketCache:  marketCache,
		locker:       locker,
		dialer:       &websocket.Dialer{HandshakeTimeout: 10 * time.Second},
		desired:      make(map[string]provider.Subscription),
		sent:         make(map[string]struct{}),
	}
}

func (s *Stream) Start(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&s.started, 0, 1) {
		return
	}
	go s.leaderLoop(ctx)
}

func (s *Stream) HasDesiredSubscriptions() bool {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	return len(s.desired) > 0
}

func (s *Stream) IsLeader() bool { return atomic.LoadInt32(&s.leader) == 1 }

func (s *Stream) SetReconnectHandler(handler func(category string)) { s.onReconnect = handler }

func (s *Stream) ReplaceSubscriptions(items []provider.Subscription) error {
	next := make(map[string]provider.Subscription)
	for _, item := range items {
		item = cache.NormalizeClientMessage(item)
		if item.CategoryCode != supportedCategory {
			continue
		}
		upstream, err := upstreamSymbol(item.Symbol)
		if err != nil {
			continue
		}
		key := canonicalSymbol(upstream)
		// One upstream price subscription feeds both the normalized quote and
		// tick topics, regardless of how many internal topics requested it.
		if current, ok := next[key]; !ok || item.Topic == types.TopicQuote || current.Topic != types.TopicQuote {
			item.Symbol = key
			next[key] = item
		}
	}
	s.subMu.Lock()
	s.desired = next
	s.subMu.Unlock()
	if atomic.LoadInt32(&s.connected) == 1 {
		return s.syncSubscriptions()
	}
	return nil
}

func (s *Stream) Resubscribe(item provider.Subscription) error {
	item = cache.NormalizeClientMessage(item)
	upstream, err := upstreamSymbol(item.Symbol)
	if err != nil {
		return err
	}
	key := canonicalSymbol(upstream)
	s.subMu.Lock()
	_, desired := s.desired[key]
	s.subMu.Unlock()
	if !desired {
		return fmt.Errorf("Twelve Data symbol is not desired: %s", key)
	}
	if atomic.LoadInt32(&s.connected) != 1 {
		return errors.New("Twelve Data websocket is not connected")
	}
	if err := s.sendSymbolAction("unsubscribe", []string{key}); err != nil {
		return err
	}
	s.subMu.Lock()
	delete(s.sent, key)
	s.subMu.Unlock()
	return s.subscribe([]string{key})
}

func (s *Stream) leaderLoop(ctx context.Context) {
	for ctx.Err() == nil {
		lockCtx, cancel := context.WithCancel(ctx)
		token, err := s.locker.acquire(lockCtx, leaderTTL)
		if err != nil {
			cancel()
			if !errors.Is(err, errLockNotObtained) {
				logx.Errorf("acquire Twelve Data leader lock failed: %v", err)
			}
			if !waitContext(ctx, 2*time.Second) {
				return
			}
			continue
		}
		atomic.StoreInt32(&s.leader, 1)
		go s.renewLoop(lockCtx, token, cancel)
		if err := s.runAsLeader(lockCtx); err != nil && !errors.Is(err, context.Canceled) {
			logx.Errorf("Twelve Data leader session stopped: %v", err)
		}
		cancel()
		s.closeConn()
		atomic.StoreInt32(&s.leader, 0)
		_ = s.locker.release(context.Background(), token)
		if !waitContext(ctx, 2*time.Second) {
			return
		}
	}
}

func (s *Stream) renewLoop(ctx context.Context, token string, cancel context.CancelFunc) {
	ticker := time.NewTicker(leaderRenewPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ok, err := s.locker.refresh(ctx, token, leaderTTL)
			if err == nil && ok {
				continue
			}
			cancel()
			s.closeConn()
			return
		}
	}
}

func (s *Stream) runAsLeader(ctx context.Context) error {
	delay := wsReconnectDelay
	for ctx.Err() == nil {
		if err := s.connect(); err != nil {
			logx.Errorf("Twelve Data websocket connect failed: %v", err)
			if !waitContext(ctx, delay+time.Duration(rand.Int63n(int64(time.Second)))) {
				return ctx.Err()
			}
			delay = min(delay*2, time.Minute)
			continue
		}
		delay = wsReconnectDelay
		atomic.StoreInt32(&s.connected, 1)
		if err := s.syncSubscriptions(); err != nil {
			logx.Errorf("restore Twelve Data subscriptions failed: %v", err)
		}
		if s.onReconnect != nil {
			go s.onReconnect(s.categoryCode)
		}
		if err := s.readLoop(ctx); err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			logx.Errorf("Twelve Data websocket read loop stopped: %v", err)
		}
		s.closeConn()
		if !waitContext(ctx, wsReconnectDelay) {
			return ctx.Err()
		}
	}
	return ctx.Err()
}

func (s *Stream) connect() error {
	endpoint, err := url.Parse(s.url)
	if err != nil {
		return err
	}
	query := endpoint.Query()
	query.Set("apikey", s.apiKey)
	endpoint.RawQuery = query.Encode()
	conn, response, err := s.dialer.Dial(endpoint.String(), http.Header{})
	if err != nil {
		errText := strings.ReplaceAll(err.Error(), s.apiKey, "[REDACTED]")
		if response != nil {
			return fmt.Errorf("dial Twelve Data websocket: status=%d: %s", response.StatusCode, errText)
		}
		return errors.New(errText)
	}
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()
	s.subMu.Lock()
	s.sent = make(map[string]struct{})
	s.subMu.Unlock()
	_ = conn.SetReadDeadline(time.Now().Add(wsReadWait))
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(wsReadWait)) })
	go s.heartbeatLoop(conn)
	logx.Infof("Twelve Data websocket connected")
	return nil
}

func (s *Stream) readLoop(ctx context.Context) error {
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		return errors.New("Twelve Data websocket connection is nil")
	}
	for ctx.Err() == nil {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		_ = conn.SetReadDeadline(time.Now().Add(wsReadWait))
		s.handleMessage(ctx, data)
	}
	return ctx.Err()
}

func (s *Stream) handleMessage(ctx context.Context, data []byte) {
	var message wsEnvelope
	if err := json.Unmarshal(data, &message); err != nil {
		logx.Errorf("decode Twelve Data websocket message failed: %v", err)
		return
	}
	switch strings.ToLower(strings.TrimSpace(message.Event)) {
	case "price":
		s.publishPrice(ctx, message)
	case "subscribe-status":
		s.handleSubscriptionACK(message)
	case "heartbeat":
		return
	default:
		if strings.EqualFold(message.Status, "error") || message.Code >= 400 {
			logx.Errorf("Twelve Data websocket protocol error: code=%d message=%s", message.Code, firstNonEmpty(message.Message, message.Status))
		}
	}
}

func (s *Stream) handleSubscriptionACK(message wsEnvelope) {
	for _, item := range message.Fails {
		key := canonicalSymbol(item.Symbol)
		s.subMu.Lock()
		delete(s.sent, key)
		s.subMu.Unlock()
		logx.Errorf("Twelve Data subscription rejected, symbol=%s code=%d reason=%s", item.Symbol, item.Code, firstNonEmpty(item.Message, item.Status))
	}
	logx.Infof("Twelve Data subscription acknowledged, accepted=%d rejected=%d status=%s", len(message.Success), len(message.Fails), message.Status)
}

func (s *Stream) publishPrice(ctx context.Context, message wsEnvelope) {
	key := canonicalSymbol(message.Symbol)
	s.subMu.Lock()
	msg, ok := s.desired[key]
	s.subMu.Unlock()
	if !ok {
		return
	}
	if message.Timestamp <= 0 {
		logx.Errorf("Twelve Data price timestamp invalid, symbol=%s", message.Symbol)
		return
	}
	priceText, price, err := rawDecimal(message.Price)
	if err != nil {
		logx.Errorf("Twelve Data price invalid, symbol=%s err=%v", message.Symbol, err)
		return
	}
	timestamp := message.Timestamp * 1000
	quoteMsg := msg
	quoteMsg.Topic = types.TopicQuote
	payload := &types.QuotePayload{LastPrice: price, LastPriceText: priceText, Volume: rawFloat(message.DayVolume), Ts: timestamp, Authority: "twelvedata-ws"}
	if err := s.marketCache.Set(ctx, quoteMsg, payload); err != nil {
		logx.Errorf("cache Twelve Data quote failed, market=%s symbol=%s err=%v", msg.Market, msg.Symbol, err)
	}
	tickMsg := msg
	tickMsg.Topic = types.TopicTick
	// day_volume is cumulative session volume, not this event's trade size. Do
	// not feed it into tick aggregation or every update would overcount volume.
	tick := &types.TickPayload{LastPrice: price, Ts: timestamp}
	if err := s.marketCache.Set(ctx, tickMsg, tick); err != nil {
		logx.Errorf("cache Twelve Data tick failed, market=%s symbol=%s err=%v", msg.Market, msg.Symbol, err)
	}
}

func (s *Stream) syncSubscriptions() error {
	s.subMu.Lock()
	toRemove := make([]string, 0)
	toAdd := make([]string, 0)
	for symbol := range s.sent {
		if _, ok := s.desired[symbol]; !ok {
			toRemove = append(toRemove, symbol)
		}
	}
	for symbol := range s.desired {
		if _, ok := s.sent[symbol]; !ok {
			toAdd = append(toAdd, symbol)
		}
	}
	s.subMu.Unlock()
	sort.Strings(toRemove)
	sort.Strings(toAdd)
	for start := 0; start < len(toRemove); start += wsSubscribeBatch {
		end := min(start+wsSubscribeBatch, len(toRemove))
		if err := s.sendSymbolAction("unsubscribe", toRemove[start:end]); err != nil {
			return err
		}
		s.subMu.Lock()
		for _, symbol := range toRemove[start:end] {
			delete(s.sent, symbol)
		}
		s.subMu.Unlock()
	}
	for start := 0; start < len(toAdd); start += wsSubscribeBatch {
		end := min(start+wsSubscribeBatch, len(toAdd))
		if err := s.subscribe(toAdd[start:end]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Stream) subscribe(symbols []string) error {
	if len(symbols) == 0 {
		return nil
	}
	if err := s.sendSymbolAction("subscribe", symbols); err != nil {
		return err
	}
	s.subMu.Lock()
	for _, symbol := range symbols {
		s.sent[canonicalSymbol(symbol)] = struct{}{}
	}
	s.subMu.Unlock()
	return nil
}

func (s *Stream) sendSymbolAction(action string, symbols []string) error {
	values := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		value, err := upstreamSymbol(symbol)
		if err != nil {
			return err
		}
		values = append(values, value)
	}
	return s.writeJSON(map[string]any{
		"action": action,
		"params": map[string]string{"symbols": strings.Join(values, ",")},
	})
}

func (s *Stream) writeJSON(value any) error {
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		return errors.New("Twelve Data websocket connection is nil")
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
	return conn.WriteJSON(value)
}

func (s *Stream) heartbeatLoop(conn *websocket.Conn) {
	ticker := time.NewTicker(wsHeartbeatPeriod)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.RLock()
		current := s.conn
		s.mu.RUnlock()
		if current != conn {
			return
		}
		if err := s.writeJSON(map[string]string{"action": "heartbeat"}); err != nil {
			s.closeConn()
			return
		}
	}
}

func (s *Stream) closeConn() {
	atomic.StoreInt32(&s.connected, 0)
	s.mu.Lock()
	conn := s.conn
	s.conn = nil
	s.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

func waitContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
