package client

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/provider"
	"wklive/services/market/internal/market/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StaleQuoteRecoveryConfig struct {
	CheckInterval time.Duration
	StaleAfter    time.Duration
	StartupGrace  time.Duration
	RestMaxAge    time.Duration
	Cooldown      time.Duration
	BatchSize     int
}

func (c StaleQuoteRecoveryConfig) withDefaults() StaleQuoteRecoveryConfig {
	if c.CheckInterval <= 0 {
		c.CheckInterval = 30 * time.Second
	}
	if c.StaleAfter <= 0 {
		c.StaleAfter = 90 * time.Second
	}
	if c.StartupGrace <= 0 {
		c.StartupGrace = 60 * time.Second
	}
	if c.RestMaxAge <= 0 {
		c.RestMaxAge = 5 * time.Minute
	}
	if c.Cooldown <= 0 {
		c.Cooldown = 60 * time.Second
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 10
	}
	return c
}

type activeProduct struct {
	ID       int64
	Category string
	Market   string
	Symbol   string
	Exchange string
}

func (p activeProduct) quoteMessage() types.ClientMessage {
	return types.ClientMessage{
		Topic: types.TopicQuote, CategoryCode: p.Category, Market: p.Market, Symbol: p.Symbol,
	}
}

func (p activeProduct) recoveryKey() string {
	return fmt.Sprintf("%s:%s:%s", p.Category, p.Market, p.Symbol)
}

func (m *MarketManager) recoverStaleQuotes(ctx context.Context) {
	ticker := time.NewTicker(m.staleConfig.CheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			m.checkStaleQuotes(ctx, now)
		}
	}
}

func (m *MarketManager) checkStaleQuotes(ctx context.Context, now time.Time) {
	if m.calendar == nil || m.source == nil || m.marketCache == nil {
		return
	}
	m.startMu.Lock()
	startedAt := m.startedAt
	m.startMu.Unlock()
	if startedAt.IsZero() || now.Sub(startedAt) < m.staleConfig.StartupGrace {
		return
	}

	products := m.activeProductSnapshot()
	openProducts := make([]activeProduct, 0, len(products))
	quoteMessages := make([]types.ClientMessage, 0, len(products))
	for _, product := range products {
		client := m.categoryClient(product.Category)
		if client == nil || !client.IsLeader() || !m.source.Supports(product.Category) {
			continue
		}
		if !m.calendar.IsProductTradingMinute(ctx, product.ID, product.Category, product.Market, product.Symbol, product.Exchange, now.UnixMilli()) {
			continue
		}
		openProducts = append(openProducts, product)
		quoteMessages = append(quoteMessages, product.quoteMessage())
	}
	if len(quoteMessages) == 0 {
		return
	}

	cached, err := m.marketCache.ReadMany(ctx, quoteMessages)
	if err != nil {
		logx.Errorf("read market quotes for stale check failed: %v", err)
		return
	}
	quotes := make(map[string]*types.QuotePayload, len(cached))
	for _, item := range cached {
		if quote, ok := item.Payload.(*types.QuotePayload); ok && quote != nil {
			quotes[cache.BuildTopicKey(item.Message)] = quote
		}
	}

	recovered := 0
	for _, product := range openProducts {
		msg := product.quoteMessage()
		quote := quotes[cache.BuildTopicKey(msg)]
		if !quoteNeedsRecovery(now, quote, m.staleConfig.StaleAfter) || !m.claimQuoteRecovery(product.recoveryKey(), now) {
			continue
		}
		m.recoverStaleQuote(ctx, now, product, quote)
		recovered++
		if recovered >= m.staleConfig.BatchSize {
			break
		}
	}
}

func (m *MarketManager) recoverStaleQuote(ctx context.Context, now time.Time, product activeProduct, cached *types.QuotePayload) {
	msg := product.quoteMessage()
	restQuote, err := m.source.FetchQuote(ctx, msg)
	if err != nil {
		logx.Errorf("stale quote REST verification failed, product_id=%d category=%s market=%s symbol=%s err=%v",
			product.ID, product.Category, product.Market, product.Symbol, err)
		return
	}
	if quoteNeedsRecovery(now, restQuote, m.staleConfig.RestMaxAge) {
		logx.Errorf("stale quote confirmed by REST, product_id=%d category=%s market=%s symbol=%s source_ts=%d",
			product.ID, product.Category, product.Market, product.Symbol, restQuote.Ts)
		return
	}
	if !restQuoteCanRecover(now, cached, restQuote, m.staleConfig.RestMaxAge) {
		logx.Infof("stale quote REST has no newer data, product_id=%d category=%s market=%s symbol=%s ws_ts=%d rest_ts=%d",
			product.ID, product.Category, product.Market, product.Symbol, cached.Ts, restQuote.Ts)
		return
	}
	if err := m.marketCache.Set(ctx, msg, restQuote); err != nil {
		logx.Errorf("publish REST-verified quote failed, product_id=%d category=%s market=%s symbol=%s err=%v",
			product.ID, product.Category, product.Market, product.Symbol, err)
		return
	}
	client := m.categoryClient(product.Category)
	if client == nil || !client.IsLeader() {
		return
	}
	if err := client.Resubscribe(msg); err != nil {
		logx.Errorf("targeted market resubscribe failed, product_id=%d category=%s market=%s symbol=%s err=%v",
			product.ID, product.Category, product.Market, product.Symbol, err)
		return
	}
	logx.Infof("targeted market resubscribe sent after REST recovery, product_id=%d category=%s market=%s symbol=%s old_ts=%d rest_ts=%d",
		product.ID, product.Category, product.Market, product.Symbol, quoteTimestamp(cached), restQuote.Ts)
}

func quoteNeedsRecovery(now time.Time, quote *types.QuotePayload, staleAfter time.Duration) bool {
	if quote == nil || quote.Ts <= 0 {
		return true
	}
	return now.UnixMilli()-quote.Ts > staleAfter.Milliseconds()
}

func restQuoteCanRecover(now time.Time, cached, rest *types.QuotePayload, maxAge time.Duration) bool {
	if quoteNeedsRecovery(now, rest, maxAge) {
		return false
	}
	return cached == nil || rest.Ts > cached.Ts
}

func quoteTimestamp(quote *types.QuotePayload) int64 {
	if quote == nil {
		return 0
	}
	return quote.Ts
}

func (m *MarketManager) claimQuoteRecovery(key string, now time.Time) bool {
	m.recoveryMu.Lock()
	defer m.recoveryMu.Unlock()
	if previous, ok := m.quoteRecoveryAt[key]; ok && now.Sub(previous) < m.staleConfig.Cooldown {
		return false
	}
	m.quoteRecoveryAt[key] = now
	return true
}

func (m *MarketManager) categoryClient(category string) provider.Stream {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[strings.ToLower(strings.TrimSpace(category))]
}

func (m *MarketManager) activeProductSnapshot() []activeProduct {
	m.productsMu.RLock()
	products := make([]activeProduct, 0, len(m.activeProducts))
	for _, product := range m.activeProducts {
		products = append(products, product)
	}
	m.productsMu.RUnlock()
	sort.Slice(products, func(i, j int) bool { return products[i].ID < products[j].ID })
	return products
}
