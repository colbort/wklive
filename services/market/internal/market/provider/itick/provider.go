// Package itick adapts iTick REST and WebSocket APIs to the vendor-neutral
// market provider contracts.
package itick

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/provider"
	"wklive/services/market/internal/pkg/itickrest"

	"github.com/redis/go-redis/v9"
)

const providerCode = "itick"

type Provider struct {
	wsURL          string
	token          string
	streamReady    bool
	restReady      bool
	marketCache    *cache.MarketDataCache
	preheater      *Preheater
	lockRedis      *redis.Client
	connectLimiter *RedisConnectLimiter
}

func New(
	wsURL string,
	apiURL string,
	token string,
	marketCache *cache.MarketDataCache,
	lockRedis *redis.Client,
	restClient *itickrest.Client,
) *Provider {
	normalizedWSUrl := strings.TrimRight(strings.TrimSpace(wsURL), "/")
	return &Provider{
		wsURL:          normalizedWSUrl,
		token:          strings.TrimSpace(token),
		streamReady:    normalizedWSUrl != "" && marketCache != nil && lockRedis != nil,
		restReady:      strings.TrimSpace(apiURL) != "" && marketCache != nil && restClient != nil,
		marketCache:    marketCache,
		preheater:      NewPreheater(apiURL, marketCache, restClient),
		lockRedis:      lockRedis,
		connectLimiter: NewRedisConnectLimiter(lockRedis),
	}
}

func (p *Provider) Code() string { return providerCode }

func (p *Provider) Supports(category string) bool {
	return p != nil && (p.streamReady || p.restReady) && p.preheater != nil && !p.preheater.IsUnsupported(category)
}

func (p *Provider) Warm(ctx context.Context, subscriptions []provider.Subscription) {
	if p == nil || !p.restReady || p.preheater == nil {
		return
	}
	p.preheater.Warm(ctx, subscriptions)
}

func (p *Provider) FetchQuote(ctx context.Context, subscription provider.Subscription) (*provider.Quote, error) {
	if p == nil || !p.restReady || p.preheater == nil {
		return nil, fmt.Errorf("%s realtime provider is not configured", providerCode)
	}
	return p.preheater.FetchQuote(ctx, subscription)
}

func (p *Provider) NewStream(category string) (provider.Stream, error) {
	if p == nil || !p.streamReady {
		return nil, fmt.Errorf("%s stream provider is not configured", providerCode)
	}
	category = strings.ToLower(strings.TrimSpace(category))
	if category == "" {
		return nil, fmt.Errorf("%s stream category is required", providerCode)
	}
	upstreamURL := p.wsURL + "/" + category
	lockKey := "market:leader:" + sha1Hex(upstreamURL)
	return NewMarketWsClient(
		upstreamURL,
		p.token,
		category,
		p.marketCache,
		NewRedisLeaderLock(p.lockRedis, lockKey),
		p.connectLimiter,
	), nil
}

func sha1Hex(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

var _ provider.RealtimeProvider = (*Provider)(nil)
var _ provider.Stream = (*ITickWsClient)(nil)
