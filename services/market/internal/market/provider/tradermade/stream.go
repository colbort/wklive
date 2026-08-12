package tradermade

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
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
	wsAuthWait        = 10 * time.Second
	wsPongWait        = 70 * time.Second
	wsPingPeriod      = 30 * time.Second
	wsReconnectDelay  = 5 * time.Second
	leaderTTL         = 15 * time.Second
	leaderRenewPeriod = 5 * time.Second
	wsSubscribeBatch  = 50
)

type Stream struct {
	url          string
	key          string
	enableLadder bool
	categoryCode string
	marketCache  *cache.MarketDataCache
	locker       *redisLeaderLock
	dialer       *websocket.Dialer

	mu      sync.RWMutex
	conn    *websocket.Conn
	writeMu sync.Mutex

	subMu       sync.Mutex
	desired     map[string]provider.Subscription
	sent        map[string]struct{}
	symbolLimit int

	leader        int32
	started       int32
	authenticated int32
	onReconnect   func(string)
}

func newStream(url, key string, enableLadder bool, marketCache *cache.MarketDataCache, locker *redisLeaderLock) *Stream {
	return &Stream{
		url:          strings.TrimSpace(url),
		key:          strings.TrimSpace(key),
		enableLadder: enableLadder,
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
		symbol := canonicalSymbol(item.Symbol)
		if symbol == "" {
			continue
		}
		// TraderMade QUOTE carries both best bid/ask and midpoint. The system's
		// multiple internal topics therefore consume one upstream symbol slot.
		if current, ok := next[symbol]; !ok || item.Topic == types.TopicQuote || current.Topic != types.TopicQuote {
			item.Symbol = symbol
			next[symbol] = item
		}
	}
	s.subMu.Lock()
	s.desired = next
	s.subMu.Unlock()
	if atomic.LoadInt32(&s.authenticated) == 1 {
		return s.syncSubscriptions()
	}
	return nil
}

func (s *Stream) Resubscribe(item provider.Subscription) error {
	item = cache.NormalizeClientMessage(item)
	symbol := canonicalSymbol(item.Symbol)
	s.subMu.Lock()
	_, desired := s.desired[symbol]
	s.subMu.Unlock()
	if !desired {
		return fmt.Errorf("TraderMade symbol is not desired: %s", symbol)
	}
	if atomic.LoadInt32(&s.authenticated) != 1 {
		return errors.New("TraderMade websocket is not authenticated")
	}
	if err := s.writeJSON(map[string]any{"action": "unsubscribe", "symbols": []string{symbol + ":QUOTE"}}); err != nil {
		return err
	}
	s.subMu.Lock()
	delete(s.sent, symbol)
	s.subMu.Unlock()
	return s.subscribe([]string{symbol})
}

func (s *Stream) leaderLoop(ctx context.Context) {
	for ctx.Err() == nil {
		lockCtx, cancel := context.WithCancel(ctx)
		token, err := s.locker.acquire(lockCtx, leaderTTL)
		if err != nil {
			cancel()
			if !errors.Is(err, errLockNotObtained) {
				logx.Errorf("acquire TraderMade leader lock failed: %v", err)
			}
			if !waitContext(ctx, 2*time.Second) {
				return
			}
			continue
		}
		atomic.StoreInt32(&s.leader, 1)
		lost := make(chan struct{}, 1)
		go s.renewLoop(lockCtx, token, lost, cancel)
		if err := s.runAsLeader(lockCtx); err != nil && !errors.Is(err, context.Canceled) {
			logx.Errorf("TraderMade leader session stopped: %v", err)
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

func (s *Stream) renewLoop(ctx context.Context, token string, lost chan<- struct{}, cancel context.CancelFunc) {
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
			select {
			case lost <- struct{}{}:
			default:
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
		if err := s.connectAndAuthenticate(); err != nil {
			logx.Errorf("TraderMade websocket connect/auth failed: %v", err)
			if !waitContext(ctx, delay+time.Duration(rand.Int63n(int64(time.Second)))) {
				return ctx.Err()
			}
			delay = min(delay*2, time.Minute)
			continue
		}
		delay = wsReconnectDelay
		atomic.StoreInt32(&s.authenticated, 1)
		if err := s.syncSubscriptions(); err != nil {
			logx.Errorf("restore TraderMade subscriptions failed: %v", err)
		}
		if s.onReconnect != nil {
			go s.onReconnect(s.categoryCode)
		}
		if err := s.readLoop(ctx); err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			logx.Errorf("TraderMade websocket read loop stopped: %v", err)
		}
		s.closeConn()
		if !waitContext(ctx, wsReconnectDelay) {
			return ctx.Err()
		}
	}
	return ctx.Err()
}

func (s *Stream) connectAndAuthenticate() error {
	conn, response, err := s.dialer.Dial(s.url, nil)
	if err != nil {
		if response != nil {
			return fmt.Errorf("dial TraderMade websocket: status=%d: %w", response.StatusCode, err)
		}
		return err
	}
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()
	s.subMu.Lock()
	s.sent = make(map[string]struct{})
	s.symbolLimit = 0
	s.subMu.Unlock()
	_ = conn.SetReadDeadline(time.Now().Add(wsAuthWait))
	login := map[string]any{"action": "login", "key": s.key, "fmt": "JSON"}
	if s.enableLadder {
		login["send_ladder"] = true
	}
	if err := s.writeJSON(login); err != nil {
		s.closeConn()
		return err
	}
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			s.closeConn()
			return err
		}
		var message wsEnvelope
		if err := json.Unmarshal(data, &message); err != nil {
			continue
		}
		switch strings.ToLower(message.Type) {
		case "login_ok":
			s.subMu.Lock()
			s.symbolLimit = message.SymbolLimit
			s.subMu.Unlock()
			_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
			conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(wsPongWait)) })
			go s.keepaliveLoop(conn)
			if s.enableLadder && !message.TraderLadder {
				logx.Errorf("TraderMade ladder requested but account capability trader_ladder=false; using best bid/ask only")
			}
			logx.Infof("TraderMade websocket authenticated, symbol_limit=%d ladder=%t", message.SymbolLimit, s.enableLadder)
			return nil
		case "login_reject", "logout":
			s.closeConn()
			return fmt.Errorf("TraderMade login rejected: %s", message.Reason)
		}
	}
}

func (s *Stream) readLoop(ctx context.Context) error {
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		return errors.New("TraderMade websocket connection is nil")
	}
	for ctx.Err() == nil {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
		s.handleMessage(ctx, data)
	}
	return ctx.Err()
}

func (s *Stream) handleMessage(ctx context.Context, data []byte) {
	var message wsEnvelope
	if err := json.Unmarshal(data, &message); err != nil {
		logx.Errorf("decode TraderMade websocket message failed: %v", err)
		return
	}
	switch strings.ToLower(message.Type) {
	case "sub_ack":
		s.handleSubscriptionACK(message)
	case "unsub_ack":
		return
	case "logout", "login_reject":
		logx.Errorf("TraderMade websocket session rejected: %s", message.Reason)
		s.closeConn()
	case "error":
		logx.Errorf("TraderMade websocket protocol error: %s", message.Reason)
	default:
		messageType := strings.ToUpper(strings.TrimSpace(message.MessageType))
		if message.Symbol != "" && (messageType == "QUOTE" || messageType == "LAST_QUOTE") {
			s.publishQuote(ctx, message)
		}
	}
}

func (s *Stream) handleSubscriptionACK(message wsEnvelope) {
	for _, item := range append(append([]string{}, message.Denied...), message.Invalid...) {
		symbol := subscriptionSymbol(item)
		s.subMu.Lock()
		delete(s.sent, symbol)
		s.subMu.Unlock()
		reason := message.DeniedReasons[item]
		logx.Errorf("TraderMade subscription rejected, symbol=%s reason=%s", symbol, reason)
	}
	logx.Infof("TraderMade subscription acknowledged, accepted=%d denied=%d invalid=%d", len(message.Accepted), len(message.Denied), len(message.Invalid))
}

func (s *Stream) publishQuote(ctx context.Context, message wsEnvelope) {
	symbol := canonicalSymbol(message.Symbol)
	s.subMu.Lock()
	msg, ok := s.desired[symbol]
	s.subMu.Unlock()
	if !ok {
		return
	}
	timestamp, err := parseWSTimestamp(message.Timestamp)
	if err != nil {
		logx.Errorf("TraderMade quote timestamp invalid, symbol=%s err=%v", symbol, err)
		return
	}
	priceText, price, err := message.midpoint()
	if err != nil {
		logx.Errorf("TraderMade quote price invalid, symbol=%s err=%v", symbol, err)
		return
	}
	quoteMsg := msg
	quoteMsg.Topic = types.TopicQuote
	payload := &types.QuotePayload{LastPrice: price, LastPriceText: priceText, Ts: timestamp, Authority: "tradermade-ws"}
	if err := s.marketCache.Set(ctx, quoteMsg, payload); err != nil {
		logx.Errorf("cache TraderMade quote failed, market=%s symbol=%s err=%v", msg.Market, symbol, err)
	}
	depthMsg := msg
	depthMsg.Topic = types.TopicDepth
	if err := s.marketCache.Set(ctx, depthMsg, websocketDepthPayload(message)); err != nil {
		logx.Errorf("cache TraderMade depth failed, market=%s symbol=%s err=%v", msg.Market, symbol, err)
	}
}

func websocketDepthPayload(message wsEnvelope) *types.DepthPayload {
	asks := ladderLevels(nil, true)
	bids := ladderLevels(nil, false)
	if message.Ladder != nil {
		asks = ladderLevels(message.Ladder.Asks, true)
		bids = ladderLevels(message.Ladder.Bids, false)
	}
	if len(asks) == 0 {
		asks = oneDepthLevel(message.Ask, message.AskVolume)
	}
	if len(bids) == 0 {
		bids = oneDepthLevel(message.Bid, message.BidVolume)
	}
	return &types.DepthPayload{Asks: asks, Bids: bids}
}

func ladderLevels(values [][]string, ask bool) []*types.DepthLevel {
	levels := make([]*types.DepthLevel, 0, len(values))
	for index, value := range values {
		if len(value) < 1 {
			continue
		}
		price, err := strconv.ParseFloat(value[0], 64)
		if err != nil || price <= 0 {
			continue
		}
		volume := float64(0)
		if len(value) > 1 {
			volume, _ = strconv.ParseFloat(value[1], 64)
		}
		levels = append(levels, &types.DepthLevel{Price: price, Volume: volume, Position: int64(index + 1), OriginVolume: volume})
	}
	if ask {
		sort.Slice(levels, func(i, j int) bool { return levels[i].Price < levels[j].Price })
	} else {
		sort.Slice(levels, func(i, j int) bool { return levels[i].Price > levels[j].Price })
	}
	return levels
}

func oneDepthLevel(priceText, volumeText string) []*types.DepthLevel {
	price, err := strconv.ParseFloat(strings.TrimSpace(priceText), 64)
	if err != nil || price <= 0 {
		return nil
	}
	volume, _ := strconv.ParseFloat(strings.TrimSpace(volumeText), 64)
	return []*types.DepthLevel{{Price: price, Volume: volume, Position: 1, OriginVolume: volume}}
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
	limit := s.symbolLimit
	s.subMu.Unlock()
	sort.Strings(toRemove)
	sort.Strings(toAdd)
	for start := 0; start < len(toRemove); start += wsSubscribeBatch {
		end := min(start+wsSubscribeBatch, len(toRemove))
		items := quoteSymbols(toRemove[start:end])
		if err := s.writeJSON(map[string]any{"action": "unsubscribe", "symbols": items}); err != nil {
			return err
		}
		s.subMu.Lock()
		for _, symbol := range toRemove[start:end] {
			delete(s.sent, symbol)
		}
		s.subMu.Unlock()
	}
	if limit > 0 && len(toAdd)+s.sentCount() > limit {
		logx.Errorf("TraderMade desired subscriptions exceed account limit, desired=%d symbol_limit=%d; upstream ACK decides accepted symbols", len(toAdd)+s.sentCount(), limit)
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
	if err := s.writeJSON(map[string]any{"action": "subscribe", "symbols": quoteSymbols(symbols), "send_last": true}); err != nil {
		return err
	}
	s.subMu.Lock()
	for _, symbol := range symbols {
		s.sent[canonicalSymbol(symbol)] = struct{}{}
	}
	s.subMu.Unlock()
	return nil
}

func (s *Stream) sentCount() int {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	return len(s.sent)
}

func quoteSymbols(symbols []string) []string {
	result := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		result = append(result, canonicalSymbol(symbol)+":QUOTE")
	}
	return result
}

func subscriptionSymbol(value string) string {
	value = strings.TrimSpace(strings.ToUpper(value))
	value = strings.TrimSuffix(value, ":QUOTE")
	return canonicalSymbol(value)
}

func (s *Stream) writeJSON(value any) error {
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		return errors.New("TraderMade websocket connection is nil")
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
	return conn.WriteJSON(value)
}

func (s *Stream) keepaliveLoop(conn *websocket.Conn) {
	ticker := time.NewTicker(wsPingPeriod)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.RLock()
		current := s.conn
		s.mu.RUnlock()
		if current != conn {
			return
		}
		s.writeMu.Lock()
		err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(wsWriteWait))
		s.writeMu.Unlock()
		if err != nil {
			s.closeConn()
			return
		}
	}
}

func (s *Stream) closeConn() {
	atomic.StoreInt32(&s.authenticated, 0)
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
