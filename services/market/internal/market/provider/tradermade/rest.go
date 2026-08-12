package tradermade

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
)

const restBatchSize = 50

type RESTClient struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	marketCache *cache.MarketDataCache
}

func NewRESTClient(baseURL, apiKey string, httpClient *http.Client, marketCache *cache.MarketDataCache) *RESTClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}
	return &RESTClient{
		baseURL:     strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:      strings.TrimSpace(apiKey),
		httpClient:  httpClient,
		marketCache: marketCache,
	}
}

func (c *RESTClient) Ready() bool {
	return c != nil && c.baseURL != "" && c.apiKey != "" && c.httpClient != nil
}

func (c *RESTClient) Warm(ctx context.Context, subscriptions []provider.Subscription) {
	products := uniqueProducts(subscriptions)
	for start := 0; start < len(products); start += restBatchSize {
		end := min(start+restBatchSize, len(products))
		quotes, timestamp, err := c.fetch(ctx, products[start:end])
		if err != nil {
			logx.Errorf("TraderMade REST warm failed, count=%d err=%v", end-start, err)
			continue
		}
		for _, msg := range products[start:end] {
			quote, ok := quotes[canonicalSymbol(msg.Symbol)]
			if !ok {
				logx.Errorf("TraderMade REST response missing symbol, market=%s symbol=%s", msg.Market, msg.Symbol)
				continue
			}
			payload, payloadErr := restQuotePayload(quote, timestamp)
			if payloadErr != nil {
				logx.Errorf("TraderMade REST quote invalid, symbol=%s err=%v", msg.Symbol, payloadErr)
				continue
			}
			quoteMsg := msg
			quoteMsg.Topic = types.TopicQuote
			if err := c.marketCache.Set(ctx, quoteMsg, payload); err != nil {
				logx.Errorf("cache TraderMade REST quote failed, symbol=%s err=%v", msg.Symbol, err)
			}
			depthMsg := msg
			depthMsg.Topic = types.TopicDepth
			if err := c.marketCache.Set(ctx, depthMsg, restDepthPayload(quote)); err != nil {
				logx.Errorf("cache TraderMade REST depth failed, symbol=%s err=%v", msg.Symbol, err)
			}
		}
	}
}

func (c *RESTClient) FetchQuote(ctx context.Context, subscription provider.Subscription) (*provider.Quote, error) {
	msg := cache.NormalizeClientMessage(subscription)
	quotes, timestamp, err := c.fetch(ctx, []provider.Subscription{msg})
	if err != nil {
		return nil, err
	}
	quote, ok := quotes[canonicalSymbol(msg.Symbol)]
	if !ok {
		return nil, fmt.Errorf("TraderMade REST response missing symbol %s", msg.Symbol)
	}
	return restQuotePayload(quote, timestamp)
}

func (c *RESTClient) fetch(ctx context.Context, subscriptions []provider.Subscription) (map[string]restQuote, int64, error) {
	if !c.Ready() {
		return nil, 0, fmt.Errorf("TraderMade REST client is not configured")
	}
	symbols := make([]string, 0, len(subscriptions))
	for _, msg := range subscriptions {
		if symbol := canonicalSymbol(msg.Symbol); symbol != "" {
			symbols = append(symbols, symbol)
		}
	}
	if len(symbols) == 0 {
		return nil, 0, fmt.Errorf("TraderMade REST symbols are empty")
	}
	endpoint, err := url.Parse(c.baseURL + "/live")
	if err != nil {
		return nil, 0, err
	}
	query := endpoint.Query()
	query.Set("currency", strings.Join(symbols, ","))
	query.Set("api_key", c.apiKey)
	endpoint.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("TraderMade REST request failed: %s", sanitizedRequestError(err, c.apiKey))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody := strings.ReplaceAll(strings.TrimSpace(string(body)), c.apiKey, "[REDACTED]")
		return nil, 0, fmt.Errorf("TraderMade REST http status=%d body=%q", resp.StatusCode, responseBody)
	}
	result, err := decodeRESTResponse(body)
	if err != nil {
		return nil, 0, err
	}
	timestamp := result.Timestamp * 1000
	if timestamp <= 0 {
		timestamp = time.Now().UnixMilli()
	}
	items := make(map[string]restQuote, len(result.Quotes))
	for _, quote := range result.Quotes {
		if symbol := quote.symbol(); symbol != "" {
			items[symbol] = quote
		}
	}
	return items, timestamp, nil
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

func restQuotePayload(quote restQuote, timestamp int64) (*types.QuotePayload, error) {
	text, price, err := quote.midpoint()
	if err != nil {
		return nil, err
	}
	return &types.QuotePayload{LastPrice: price, LastPriceText: text, Ts: timestamp, Authority: "tradermade-rest"}, nil
}

func restDepthPayload(quote restQuote) *types.DepthPayload {
	asks := make([]*types.DepthLevel, 0, 1)
	bids := make([]*types.DepthLevel, 0, 1)
	if value, err := quote.Ask.Float64(); err == nil && value > 0 {
		asks = append(asks, &types.DepthLevel{Price: value})
	}
	if value, err := quote.Bid.Float64(); err == nil && value > 0 {
		bids = append(bids, &types.DepthLevel{Price: value})
	}
	return &types.DepthPayload{Asks: asks, Bids: bids}
}

func uniqueProducts(subscriptions []provider.Subscription) []provider.Subscription {
	seen := make(map[string]struct{})
	result := make([]provider.Subscription, 0, len(subscriptions))
	for _, item := range subscriptions {
		item = cache.NormalizeClientMessage(item)
		if item.CategoryCode != supportedCategory || item.Symbol == "" {
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
