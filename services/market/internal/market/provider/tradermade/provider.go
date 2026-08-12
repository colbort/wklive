// Package tradermade adapts TraderMade REST and WebSocket APIs to the common
// realtime market provider boundary.
package tradermade

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/provider"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	providerCode      = "tradermade"
	supportedCategory = "forex"
)

type Provider struct {
	wsURL        string
	streamingKey string
	enableLadder bool
	marketCache  *cache.MarketDataCache
	lockRedis    *redis.Client
	restClient   *RESTClient
	streamReady  bool
	restReady    bool
}

func New(
	apiURL string,
	wsURL string,
	apiKey string,
	streamingKey string,
	enableLadder bool,
	marketCache *cache.MarketDataCache,
	lockRedis *redis.Client,
	httpClient *http.Client,
) *Provider {
	streamingKey = strings.TrimSpace(streamingKey)
	apiKey = strings.TrimSpace(apiKey)
	if streamingKey == "" || apiKey == "" {
		logx.Errorf("TraderMade provider is not configured because APIKey or StreamingAPIKey is empty")
		return nil
	}
	restClient := NewRESTClient(apiURL, apiKey, httpClient, marketCache)
	wsURL = strings.TrimSpace(wsURL)
	return &Provider{
		wsURL:        wsURL,
		streamingKey: streamingKey,
		enableLadder: enableLadder,
		marketCache:  marketCache,
		lockRedis:    lockRedis,
		restClient:   restClient,
		streamReady:  wsURL != "" && streamingKey != "" && marketCache != nil && lockRedis != nil,
		restReady:    restClient.Ready() && marketCache != nil,
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
	lockHash := sha1.Sum([]byte(p.wsURL + ":" + supportedCategory))
	lockKey := "market:leader:" + hex.EncodeToString(lockHash[:])
	return newStream(p.wsURL, p.streamingKey, p.enableLadder, p.marketCache, newRedisLeaderLock(p.lockRedis, lockKey)), nil
}

var _ provider.RealtimeProvider = (*Provider)(nil)
var _ provider.Stream = (*Stream)(nil)
