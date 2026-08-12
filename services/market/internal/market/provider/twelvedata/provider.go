// Package twelvedata adapts Twelve Data REST and WebSocket APIs to the common
// realtime market provider boundary.
package twelvedata

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/provider"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/time/rate"
)

const (
	providerCode      = "twelvedata"
	supportedCategory = "forex"
)

type Provider struct {
	wsURL       string
	apiKey      string
	marketCache *cache.MarketDataCache
	lockRedis   *redis.Client
	restClient  *RESTClient
	catalog     *SymbolCatalog
	streamReady bool
	restReady   bool
}

func New(
	apiURL string,
	wsURL string,
	apiKey string,
	restLimiter *rate.Limiter,
	warmMaxSymbols int,
	marketCache *cache.MarketDataCache,
	lockRedis *redis.Client,
	httpClient *http.Client,
) *Provider {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		logx.Errorf("Twelve Data provider is not configured because APIKey is empty")
		return nil
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}
	catalog := newSymbolCatalog(apiURL, apiKey, restLimiter, httpClient)
	restClient := newRESTClient(apiURL, apiKey, restLimiter, warmMaxSymbols, httpClient, marketCache, catalog)
	wsURL = strings.TrimSpace(wsURL)
	return &Provider{
		wsURL:       wsURL,
		apiKey:      apiKey,
		marketCache: marketCache,
		lockRedis:   lockRedis,
		restClient:  restClient,
		catalog:     catalog,
		streamReady: wsURL != "" && apiKey != "" && marketCache != nil && lockRedis != nil,
		restReady:   restClient.Ready() && marketCache != nil,
	}
}

func (p *Provider) Code() string { return providerCode }

func (p *Provider) Supports(category string) bool {
	return p != nil && strings.EqualFold(strings.TrimSpace(category), supportedCategory) && (p.streamReady || p.restReady)
}

func (p *Provider) Warm(ctx context.Context, subscriptions []provider.Subscription) {
	if p == nil || !p.restReady {
		return
	}
	p.restClient.Warm(ctx, subscriptions)
}

func (p *Provider) FetchQuote(ctx context.Context, subscription provider.Subscription) (*provider.Quote, error) {
	if p == nil || !p.restReady {
		return nil, fmt.Errorf("%s REST provider is not configured", providerCode)
	}
	return p.restClient.FetchQuote(ctx, subscription)
}

func (p *Provider) NewStream(category string) (provider.Stream, error) {
	if p == nil || !p.streamReady {
		return nil, fmt.Errorf("%s stream provider is not configured", providerCode)
	}
	if !strings.EqualFold(strings.TrimSpace(category), supportedCategory) {
		return nil, fmt.Errorf("%s does not support stream category %q", providerCode, category)
	}
	// One leader owns one batched Twelve Data connection for all forex symbols.
	// This avoids consuming one of the account-wide connection slots per pair.
	lockHash := sha1.Sum([]byte(p.wsURL + ":" + supportedCategory))
	lockKey := "market:leader:" + hex.EncodeToString(lockHash[:])
	return newStream(p.wsURL, p.apiKey, p.marketCache, p.catalog, newRedisLeaderLock(p.lockRedis, lockKey)), nil
}

var _ provider.RealtimeProvider = (*Provider)(nil)
var _ provider.Stream = (*Stream)(nil)
