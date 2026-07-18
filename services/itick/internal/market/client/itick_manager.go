package client

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"wklive/services/itick/internal/market/cache"
	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/internal/pkg/itickrest"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/models"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type ItickManager struct {
	wsUrl string
	token string

	model        models.TItickCategoryModel
	productModel models.TItickProductModel

	busRedis    *redis.Client
	lockRedis   *redis.Client
	marketCache *cache.MarketDataCache
	preheater   *cache.MarketDataPreheater

	mu      sync.RWMutex
	clients map[string]*ItickWsClient

	startMu sync.Mutex
	started bool
	runCtx  context.Context

	recoveryMu      sync.Mutex
	recoveryRunning map[string]bool
	onReconnect     func(string)
}

// LoadActiveProductSubscriptions loads the deduplicated product set maintained
// by tenant-product changes. These subscriptions are service-owned and do not
// disappear when an app websocket disconnects.
func (m *ItickManager) LoadActiveProductSubscriptions(ctx context.Context) error {
	return m.refreshActiveProductSubscriptions(ctx, true)
}

func (m *ItickManager) RefreshActiveProductSubscriptions(ctx context.Context) error {
	return m.refreshActiveProductSubscriptions(ctx, false)
}

func (m *ItickManager) refreshActiveProductSubscriptions(ctx context.Context, warm bool) error {
	if err := m.rebuildActiveProducts(ctx); err != nil {
		return fmt.Errorf("rebuild active itick products: %w", err)
	}
	ids, err := m.busRedis.SMembers(ctx, "itick:v1:active_products").Result()
	if err != nil {
		return err
	}

	msgs := make([]types.ClientMessage, 0, len(ids)*(3+len(utils.KlineIntervals)))
	for _, rawID := range ids {
		id, err := strconv.ParseInt(strings.TrimSpace(rawID), 10, 64)
		if err != nil || id <= 0 {
			logx.Errorf("skip invalid active itick product id=%q", rawID)
			continue
		}
		meta, err := m.busRedis.HGetAll(ctx, fmt.Sprintf("itick:v1:product:%d", id)).Result()
		if err != nil {
			return err
		}
		category := strings.ToLower(strings.TrimSpace(meta["category_code"]))
		market := strings.ToUpper(strings.TrimSpace(meta["market"]))
		symbol := strings.ToUpper(strings.TrimSpace(meta["symbol"]))
		if category == "" || market == "" || symbol == "" {
			logx.Errorf("skip active itick product without metadata, id=%d", id)
			continue
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

	if warm {
		m.preheater.Warm(ctx, msgs)
	}
	byCategory := make(map[string]map[string]types.ClientMessage)
	for _, msg := range normalizeUniqueMessages(msgs) {
		if m.preheater.IsUnsupported(msg.CategoryCode) {
			continue
		}
		if byCategory[msg.CategoryCode] == nil {
			byCategory[msg.CategoryCode] = make(map[string]types.ClientMessage)
		}
		byCategory[msg.CategoryCode][cache.BuildTopicKey(msg)] = msg
	}
	m.mu.RLock()
	clients := make(map[string]*ItickWsClient, len(m.clients))
	for category, cli := range m.clients {
		clients[category] = cli
	}
	m.mu.RUnlock()

	var syncErrors []error
	for category, cli := range clients {
		items := byCategory[category]
		if err := cli.replaceDesiredSubscriptions(items); err != nil && cli.IsLeader() {
			syncErr := fmt.Errorf("sync active subscriptions, category=%s: %w", category, err)
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
	logx.Infof("loaded active itick product subscriptions, products=%d topics=%d", len(ids), len(msgs))
	return errors.Join(syncErrors...)
}

func (m *ItickManager) rebuildActiveProducts(ctx context.Context) error {
	const activeKey = "itick:v1:active_products"
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
			pipe.HSet(ctx, fmt.Sprintf("itick:v1:product:%d", product.Id), map[string]any{
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
	logx.Infof("rebuilt active itick products, count=%d", count)
	return nil
}

func NewItickManager(
	wsUrl string,
	apiURL string,
	token string,
	model models.TItickCategoryModel,
	productModel models.TItickProductModel,
	busRedis *redis.Client,
	lockRedis *redis.Client,
	marketCache *cache.MarketDataCache,
	restClient *itickrest.Client,
) *ItickManager {
	return &ItickManager{
		wsUrl:           wsUrl,
		token:           token,
		model:           model,
		productModel:    productModel,
		busRedis:        busRedis,
		lockRedis:       lockRedis,
		marketCache:     marketCache,
		preheater:       cache.NewMarketDataPreheater(apiURL, marketCache, restClient),
		clients:         make(map[string]*ItickWsClient),
		recoveryRunning: make(map[string]bool),
	}
}

func (m *ItickManager) Load(ctx context.Context) error {
	categories, err := m.model.FindAll(ctx)
	if err != nil {
		return err
	}

	newClients := make(map[string]*ItickWsClient)
	connectLimiter := NewRedisConnectLimiter(m.lockRedis)

	for _, item := range categories {
		categoryCode := strings.ToLower(strings.TrimSpace(item.CategoryCode))
		wsURL := strings.TrimSpace(m.wsUrl)

		if categoryCode == "" || wsURL == "" {
			logx.Errorf("skip invalid itick category, code=%s, wsURL=%s", item.CategoryCode, m.wsUrl)
			continue
		}

		upstreamURL := fmt.Sprintf("%s/%s", wsURL, categoryCode)
		lockKey := "itick:leader:" + sha1Hex(upstreamURL)

		newClients[categoryCode] = NewItickWsClient(
			upstreamURL,
			m.token,
			categoryCode,
			m.marketCache,
			NewRedisLeaderLock(m.lockRedis, lockKey),
			connectLimiter,
		)
		newClients[categoryCode].SetReconnectHandler(m.handleReconnect)
	}

	if len(newClients) == 0 {
		return fmt.Errorf("no available itick categories found")
	}

	m.mu.Lock()
	m.clients = newClients
	m.mu.Unlock()

	logx.Infof("itick manager loaded categories success, count=%d", len(newClients))
	return nil
}

func (m *ItickManager) SetReconnectHandler(handler func(category string)) {
	m.recoveryMu.Lock()
	m.onReconnect = handler
	m.recoveryMu.Unlock()
}

func (m *ItickManager) handleReconnect(category string) {
	m.recoveryMu.Lock()
	if m.recoveryRunning[category] || m.onReconnect == nil {
		m.recoveryMu.Unlock()
		return
	}
	m.recoveryRunning[category] = true
	handler := m.onReconnect
	m.recoveryMu.Unlock()
	defer func() {
		m.recoveryMu.Lock()
		delete(m.recoveryRunning, category)
		m.recoveryMu.Unlock()
	}()
	handler(category)
}

func (m *ItickManager) Start(ctx context.Context) error {
	m.startMu.Lock()
	if m.started {
		m.startMu.Unlock()
		return nil
	}
	m.started = true
	m.runCtx = ctx
	m.startMu.Unlock()

	m.mu.RLock()
	clients := make([]*ItickWsClient, 0, len(m.clients))
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

	return nil
}

func (m *ItickManager) reconcileActiveProducts(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.RefreshActiveProductSubscriptions(ctx); err != nil {
				logx.Errorf("reconcile active itick products failed: %v", err)
			}
		}
	}
}

func (m *ItickManager) SetQuoteHandler(handler func(ctx context.Context, msg types.ClientMessage, payload *types.QuotePayload)) {
	m.marketCache.SetQuoteHandler(handler)
}

func sha1Hex(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
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
