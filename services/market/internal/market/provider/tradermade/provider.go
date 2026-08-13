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
	"time"

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
	wsURL         string
	streamingKey  string
	enableLadder  bool
	marketCache   *cache.MarketDataCache
	lockRedis     *redis.Client
	restClient    *RESTClient
	streamCatalog *SymbolCatalog
	streamReady   bool
	restReady     bool
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
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}
	restCatalog := newRESTSymbolCatalog(apiURL, apiKey, httpClient)
	streamCatalog := newStreamSymbolCatalog(apiURL, apiKey, httpClient)
	restClient := newRESTClient(apiURL, apiKey, httpClient, marketCache, restCatalog)
	wsURL = strings.TrimSpace(wsURL)
	return &Provider{
		wsURL:         wsURL,
		streamingKey:  streamingKey,
		enableLadder:  enableLadder,
		marketCache:   marketCache,
		lockRedis:     lockRedis,
		restClient:    restClient,
		streamCatalog: streamCatalog,
		streamReady:   wsURL != "" && streamingKey != "" && marketCache != nil && lockRedis != nil,
		restReady:     restClient.Ready() && marketCache != nil,
	}
}

func (p *Provider) Code() string { return providerCode }

func (p *Provider) Categories() []string { return []string{supportedCategory} }

func (p *Provider) Supports(category string) bool {
	return p != nil && strings.EqualFold(strings.TrimSpace(category), supportedCategory) && (p.streamReady || p.restReady)
}

func (p *Provider) Warm(ctx context.Context, subscriptions []provider.Subscription) {
	if p == nil {
		return
	}
	if p.restReady {
		p.restClient.Warm(ctx, subscriptions)
	}
	if p.streamReady {
		if err := p.streamCatalog.Load(ctx); err != nil {
			logx.Errorf("load TraderMade streaming symbol catalog failed: %v", err)
		} else {
			logx.Infof("loaded TraderMade streaming symbol catalog, count=%d", p.streamCatalog.Count())
		}
	}
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
	return newStream(p.wsURL, p.streamingKey, p.enableLadder, p.marketCache, p.streamCatalog, newRedisLeaderLock(p.lockRedis, lockKey)), nil
}

var _ provider.RealtimeProvider = (*Provider)(nil)
var _ provider.Stream = (*Stream)(nil)
