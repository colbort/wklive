package client

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/calendar"
	"wklive/services/market/internal/market/provider"
	"wklive/services/market/internal/market/types"
	"wklive/services/market/internal/pkg/utils"
	"wklive/services/market/models"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type MarketManager struct {
	source provider.RealtimeProvider

	productModel models.TItickProductModel

	busRedis    *redis.Client
	marketCache *cache.MarketDataCache
	calendar    *calendar.Resolver
	staleConfig StaleQuoteRecoveryConfig

	mu      sync.RWMutex
	clients map[streamKey]provider.Stream

	startMu   sync.Mutex
	started   bool
	runCtx    context.Context
	startedAt time.Time

	productsMu     sync.RWMutex
	activeProducts map[int64]activeProduct

	recoveryMu      sync.Mutex
	recoveryRunning map[streamKey]bool
	quoteRecoveryAt map[string]time.Time
	onReconnect     func(providerCode, category string)
}

type streamKey struct {
	provider string
	category string
}

// LoadActiveProductSubscriptions loads the deduplicated product set maintained
// by tenant-product changes. These subscriptions are service-owned and do not
// disappear when an app websocket disconnects.
func (m *MarketManager) LoadActiveProductSubscriptions(ctx context.Context) error {
	return m.refreshActiveProductSubscriptions(ctx, true)
}

func (m *MarketManager) RefreshActiveProductSubscriptions(ctx context.Context) error {
	return m.refreshActiveProductSubscriptions(ctx, false)
}

func (m *MarketManager) refreshActiveProductSubscriptions(ctx context.Context, warm bool) error {
	if err := m.rebuildActiveProducts(ctx); err != nil {
		return fmt.Errorf("rebuild active market products: %w", err)
	}
	ids, err := m.busRedis.SMembers(ctx, "market:v1:active_products").Result()
	if err != nil {
		return err
	}

	msgs := make([]types.ClientMessage, 0, len(ids)*(3+len(utils.KlineIntervals)))
	activeProducts := make(map[int64]activeProduct, len(ids))
	for _, rawID := range ids {
		id, err := strconv.ParseInt(strings.TrimSpace(rawID), 10, 64)
		if err != nil || id <= 0 {
			logx.Errorf("skip invalid active market product id=%q", rawID)
			continue
		}
		meta, err := m.busRedis.HGetAll(ctx, fmt.Sprintf("market:v1:product:%d", id)).Result()
		if err != nil {
			return err
		}
		category := strings.ToLower(strings.TrimSpace(meta["category_code"]))
		market := strings.ToUpper(strings.TrimSpace(meta["market"]))
		symbol := strings.ToUpper(strings.TrimSpace(meta["symbol"]))
		exchange := strings.TrimSpace(meta["exchange"])
		if category == "" || market == "" || symbol == "" {
			logx.Errorf("skip active market product without metadata, id=%d", id)
			continue
		}
		activeProducts[id] = activeProduct{
			ID: id, Category: category, Market: market, Symbol: symbol, Exchange: exchange,
		}

		for _, topic := range []types.Topic{types.TopicDepth, types.TopicTick, types.TopicQuote} {
			msgs = append(msgs, types.ClientMessage{Topic: topic, CategoryCode: category, Market: market, Symbol: symbol})
		}
		for _, interval := range utils.KlineIntervals {
			msgs = append(msgs, types.ClientMessage{
				Topic: types.TopicKline, CategoryCode: category, Market: market,
				Symbol: symbol, Interval: interval.Name,
			})
		}
	}
	m.productsMu.Lock()
	m.activeProducts = activeProducts
	m.productsMu.Unlock()

	if m.source == nil {
		return errors.New("realtime market provider is not configured")
	}
	if warm {
		m.source.Warm(ctx, msgs)
	}
	byCategory := make(map[string]map[string]types.ClientMessage)
	for _, msg := range normalizeUniqueMessages(msgs) {
		if !m.source.Supports(msg.CategoryCode) {
			continue
		}
		if byCategory[msg.CategoryCode] == nil {
			byCategory[msg.CategoryCode] = make(map[string]types.ClientMessage)
		}
		byCategory[msg.CategoryCode][cache.BuildTopicKey(msg)] = msg
	}
	m.mu.RLock()
	clients := make(map[streamKey]provider.Stream, len(m.clients))
	for key, cli := range m.clients {
		clients[key] = cli
	}
	m.mu.RUnlock()

	var syncErrors []error
	for key, cli := range clients {
		items := byCategory[key.category]
		if err := cli.ReplaceSubscriptions(subscriptionValues(items)); err != nil && cli.IsLeader() {
			syncErr := fmt.Errorf("sync active subscriptions, provider=%s category=%s: %w", key.provider, key.category, err)
			logx.Errorf("%v", syncErr)
			syncErrors = append(syncErrors, syncErr)
		}
		m.startMu.Lock()
		started, runCtx := m.started, m.runCtx
		m.startMu.Unlock()
		if started && len(items) > 0 && runCtx != nil {
			cli.Start(runCtx)
		}
	}
	logx.Infof("loaded active market product subscriptions, products=%d topics=%d", len(ids), len(msgs))
	return errors.Join(syncErrors...)
}

func (m *MarketManager) rebuildActiveProducts(ctx context.Context) error {
	const activeKey = "market:v1:active_products"
	tempKey := fmt.Sprintf("%s:rebuild:%d", activeKey, time.Now().UnixNano())
	// The marker keeps the temporary set alive when there are zero products.
	if err := m.busRedis.SAdd(ctx, tempKey, "__empty__").Err(); err != nil {
		return err
	}
	defer m.busRedis.Del(context.Background(), tempKey)

	var cursor int64
	var count int
	for {
		products, err := m.productModel.FindActivePage(ctx, cursor, 500)
		if err != nil {
			return err
		}
		if len(products) == 0 {
			break
		}
		pipe := m.busRedis.Pipeline()
		for _, product := range products {
			pipe.SAdd(ctx, tempKey, product.Id)
			pipe.HSet(ctx, fmt.Sprintf("market:v1:product:%d", product.Id), map[string]any{
				"category_code": product.CategoryCode,
				"market":        product.Market,
				"symbol":        product.Symbol,
				"exchange":      product.Exchange,
			})
			cursor = product.Id
			count++
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
		if len(products) < 500 {
			break
		}
	}
	if err := m.busRedis.Rename(ctx, tempKey, activeKey).Err(); err != nil {
		return err
	}
	if err := m.busRedis.SRem(ctx, activeKey, "__empty__").Err(); err != nil {
		return err
	}
	logx.Infof("rebuilt active market products, count=%d", count)
	return nil
}

func NewMarketManager(
	source provider.RealtimeProvider,
	productModel models.TItickProductModel,
	busRedis *redis.Client,
	marketCache *cache.MarketDataCache,
	calendarResolver *calendar.Resolver,
	staleConfig StaleQuoteRecoveryConfig,
) *MarketManager {
	return &MarketManager{
		source:          source,
		productModel:    productModel,
		busRedis:        busRedis,
		marketCache:     marketCache,
		calendar:        calendarResolver,
		staleConfig:     staleConfig.withDefaults(),
		clients:         make(map[streamKey]provider.Stream),
		activeProducts:  make(map[int64]activeProduct),
		recoveryRunning: make(map[streamKey]bool),
		quoteRecoveryAt: make(map[string]time.Time),
	}
}

func (m *MarketManager) Load(ctx context.Context) error {
	if m.source == nil {
		return errors.New("realtime market provider is not configured")
	}
	newClients := make(map[streamKey]provider.Stream)

	for _, source := range provider.Sources(m.source) {
		providerCode := strings.ToLower(strings.TrimSpace(source.Code()))
		if providerCode == "" {
			continue
		}
		seenCategories := make(map[string]struct{})
		for _, category := range source.Categories() {
			category = strings.ToLower(strings.TrimSpace(category))
			if category == "" {
				continue
			}
			if _, exists := seenCategories[category]; exists {
				continue
			}
			seenCategories[category] = struct{}{}
			if !source.Supports(category) {
				logx.Errorf("skip unsupported market category, provider=%s category=%s", providerCode, category)
				continue
			}
			stream, err := source.NewStream(category)
			if err != nil {
				logx.Errorf("create market stream failed, provider=%s category=%s err=%v", providerCode, category, err)
				continue
			}
			key := streamKey{provider: providerCode, category: category}
			stream.SetReconnectHandler(func(string) { m.handleReconnect(key) })
			newClients[key] = stream
		}
	}

	if len(newClients) == 0 {
		return fmt.Errorf("no available market categories found")
	}

	m.mu.Lock()
	m.clients = newClients
	m.mu.Unlock()

	logx.Infof("market manager loaded provider streams success, providers=%d streams=%d", len(provider.Sources(m.source)), len(newClients))
	return nil
}

func (m *MarketManager) SetReconnectHandler(handler func(providerCode, category string)) {
	m.recoveryMu.Lock()
	m.onReconnect = handler
	m.recoveryMu.Unlock()
}

func (m *MarketManager) handleReconnect(key streamKey) {
	m.recoveryMu.Lock()
	if m.recoveryRunning[key] || m.onReconnect == nil {
		m.recoveryMu.Unlock()
		return
	}
	m.recoveryRunning[key] = true
	handler := m.onReconnect
	m.recoveryMu.Unlock()
	defer func() {
		m.recoveryMu.Lock()
		delete(m.recoveryRunning, key)
		m.recoveryMu.Unlock()
	}()
	handler(key.provider, key.category)
}

func (m *MarketManager) Start(ctx context.Context) error {
	m.startMu.Lock()
	if m.started {
		m.startMu.Unlock()
		return nil
	}
	m.started = true
	m.runCtx = ctx
	m.startedAt = time.Now()
	m.startMu.Unlock()

	m.mu.RLock()
	clients := make([]provider.Stream, 0, len(m.clients))
	for _, cli := range m.clients {
		clients = append(clients, cli)
	}
	m.mu.RUnlock()

	for _, cli := range clients {
		if cli.HasDesiredSubscriptions() {
			cli.Start(ctx)
		}
	}
	go m.reconcileActiveProducts(ctx)
	go m.recoverStaleQuotes(ctx)

	return nil
}

func (m *MarketManager) reconcileActiveProducts(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.RefreshActiveProductSubscriptions(ctx); err != nil {
				logx.Errorf("reconcile active market products failed: %v", err)
			}
		}
	}
}

func (m *MarketManager) SetQuoteHandler(handler func(ctx context.Context, msg types.ClientMessage, payload *types.QuotePayload) error) {
	m.marketCache.SetQuoteHandler(handler)
}

func normalizeUniqueMessages(msgs []types.ClientMessage) []types.ClientMessage {
	out := make([]types.ClientMessage, 0, len(msgs))
	seen := make(map[string]struct{}, len(msgs))
	for _, msg := range msgs {
		msg = cache.NormalizeClientMessage(msg)
		key := cache.BuildTopicKey(msg)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, msg)
	}
	return out
}

func subscriptionValues(items map[string]types.ClientMessage) []types.ClientMessage {
	out := make([]types.ClientMessage, 0, len(items))
	for _, msg := range items {
		out = append(out, msg)
	}
	sort.Slice(out, func(i, j int) bool {
		return cache.BuildTopicKey(out[i]) < cache.BuildTopicKey(out[j])
	})
	return out
}
