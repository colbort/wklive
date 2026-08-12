package twelvedata

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/provider"
	"wklive/services/market/internal/market/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/time/rate"
)

type RESTClient struct {
	baseURL     string
	apiKey      string
	limiter     *rate.Limiter
	warmMax     int
	httpClient  *http.Client
	marketCache *cache.MarketDataCache
	catalog     *SymbolCatalog
}

func NewRESTClient(baseURL, apiKey string, limiter *rate.Limiter, warmMax int, httpClient *http.Client, marketCache *cache.MarketDataCache) *RESTClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}
	catalog := newSymbolCatalog(baseURL, apiKey, limiter, httpClient)
	return newRESTClient(baseURL, apiKey, limiter, warmMax, httpClient, marketCache, catalog)
}

func newRESTClient(baseURL, apiKey string, limiter *rate.Limiter, warmMax int, httpClient *http.Client, marketCache *cache.MarketDataCache, catalog *SymbolCatalog) *RESTClient {
	return &RESTClient{
		baseURL:     strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:      strings.TrimSpace(apiKey),
		limiter:     limiter,
		warmMax:     warmMax,
		httpClient:  httpClient,
		marketCache: marketCache,
		catalog:     catalog,
	}
}

func (c *RESTClient) Ready() bool {
	return c != nil && c.baseURL != "" && c.apiKey != "" && c.httpClient != nil
}

func (c *RESTClient) Warm(ctx context.Context, subscriptions []provider.Subscription) {
	if err := c.catalog.Load(ctx); err != nil {
		logx.Errorf("load Twelve Data forex symbol catalog failed: %v", err)
		return
	}
	logx.Infof("loaded Twelve Data forex symbol catalog, count=%d", c.catalog.Count())
	products := uniqueProducts(subscriptions, c.catalog)
	if c.warmMax > 0 && len(products) > c.warmMax {
		logx.Infof("Twelve Data REST warm limited, total=%d selected=%d", len(products), c.warmMax)
		products = products[:c.warmMax]
	}
	for _, msg := range products {
		payload, err := c.FetchQuote(ctx, msg)
		if err != nil {
			logx.Errorf("Twelve Data REST warm failed, market=%s symbol=%s err=%v", msg.Market, msg.Symbol, err)
			continue
		}
		msg.Topic = types.TopicQuote
		if err := c.marketCache.Set(ctx, msg, payload); err != nil {
			logx.Errorf("cache Twelve Data REST quote failed, symbol=%s err=%v", msg.Symbol, err)
		}
	}
}

func (c *RESTClient) FetchQuote(ctx context.Context, subscription provider.Subscription) (*provider.Quote, error) {
	if !c.Ready() {
		return nil, fmt.Errorf("Twelve Data REST client is not configured")
	}
	msg := cache.NormalizeClientMessage(subscription)
	if err := c.catalog.Load(ctx); err != nil {
		return nil, err
	}
	symbol, err := c.catalog.Resolve(msg.Symbol)
	if err != nil {
		return nil, err
	}
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("wait for Twelve Data REST rate limit: %w", err)
		}
	}
	endpoint, err := url.Parse(c.baseURL + "/quote")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("symbol", symbol)
	query.Set("interval", "1min")
	query.Set("timezone", "UTC")
	endpoint.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	// Twelve Data recommends header authentication; this also keeps the key out
	// of request URLs and transport errors.
	req.Header.Set("Authorization", "apikey "+c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Twelve Data REST request failed: %s", sanitizedRequestError(err, c.apiKey))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody := strings.ReplaceAll(strings.TrimSpace(string(body)), c.apiKey, "[REDACTED]")
		return nil, fmt.Errorf("Twelve Data REST http status=%d body=%q", resp.StatusCode, responseBody)
	}
	quote, err := decodeRESTQuote(body)
	if err != nil {
		return nil, err
	}
	if canonicalSymbol(quote.Symbol) != canonicalSymbol(msg.Symbol) {
		return nil, fmt.Errorf("Twelve Data REST response symbol mismatch: requested=%s received=%s", msg.Symbol, quote.Symbol)
	}
	text, price, err := quote.price()
	if err != nil {
		return nil, err
	}
	return &types.QuotePayload{
		LastPrice:     price,
		LastPriceText: text,
		Ts:            quote.Timestamp * 1000,
		Authority:     "twelvedata-rest",
	}, nil
}

func sanitizedRequestError(err error, apiKey string) string {
	current := err
	for {
		var urlError *url.Error
		if !errors.As(current, &urlError) || urlError == nil || urlError.Err == nil {
			break
		}
		current = urlError.Err
	}
	return strings.ReplaceAll(current.Error(), apiKey, "[REDACTED]")
}

func uniqueProducts(subscriptions []provider.Subscription, catalog *SymbolCatalog) []provider.Subscription {
	seen := make(map[string]struct{})
	result := make([]provider.Subscription, 0, len(subscriptions))
	for _, item := range subscriptions {
		item = cache.NormalizeClientMessage(item)
		if item.CategoryCode != supportedCategory || item.Symbol == "" {
			continue
		}
		if _, err := catalog.Resolve(item.Symbol); err != nil {
			continue
		}
		key := item.CategoryCode + ":" + item.Market + ":" + canonicalSymbol(item.Symbol)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}
