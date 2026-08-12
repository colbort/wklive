package itick

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"

	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/types"
	"wklive/services/market/internal/pkg/itickrest"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

type Preheater struct {
	apiURL      string
	cache       *cache.MarketDataCache
	restClient  *itickrest.Client
	mu          sync.RWMutex
	unsupported map[string]struct{}
}

func NewPreheater(apiURL string, cache *cache.MarketDataCache, restClient *itickrest.Client) *Preheater {
	return &Preheater{
		apiURL:      strings.TrimRight(strings.TrimSpace(apiURL), "/"),
		cache:       cache,
		restClient:  restClient,
		unsupported: make(map[string]struct{}),
	}
}

const marketDataPreheatBatchSize = 10

type marketDataBatch struct {
	category string
	market   string
	topic    types.Topic
	msgs     []types.ClientMessage
}

// Warm fetches REST snapshots through the batch quotes/ticks/depths endpoints.
// Kline messages are ignored because their reconciliation has a separate flow.
func (p *Preheater) Warm(ctx context.Context, msgs []types.ClientMessage) {
	groups := make(map[string][]types.ClientMessage)
	seen := make(map[string]struct{})
	for _, msg := range msgs {
		msg = cache.NormalizeClientMessage(msg)
		if msg.Topic != types.TopicQuote && msg.Topic != types.TopicTick && msg.Topic != types.TopicDepth {
			continue
		}
		topicKey := cache.BuildTopicKey(msg)
		if _, ok := seen[topicKey]; ok {
			continue
		}
		seen[topicKey] = struct{}{}
		groupKey := msg.CategoryCode + ":" + msg.Market + ":" + string(msg.Topic)
		groups[groupKey] = append(groups[groupKey], msg)
	}

	batchesByCategory := make(map[string][]marketDataBatch)
	for _, items := range groups {
		for start := 0; start < len(items); start += marketDataPreheatBatchSize {
			end := min(start+marketDataPreheatBatchSize, len(items))
			batch := marketDataBatch{category: items[start].CategoryCode, market: items[start].Market,
				topic: items[start].Topic, msgs: items[start:end]}
			batchesByCategory[batch.category] = append(batchesByCategory[batch.category], batch)
		}
	}

	mr.ForEach(func(source chan<- []marketDataBatch) {
		for _, batches := range batchesByCategory {
			source <- batches
		}
	}, func(batches []marketDataBatch) {
		for _, batch := range batches {
			err := p.fetchBatchAndCache(ctx, batch)
			if err == nil {
				continue
			}
			if isPackageUnsupported(err) {
				p.markUnsupported(batch.category)
				logx.Errorf("market package does not support category, category=%s err=%v", batch.category, err)
				break
			}
			logx.Errorf("preheat market market data batch failed, topic=%s category=%s market=%s count=%d err=%v",
				batch.topic, batch.category, batch.market, len(batch.msgs), err)
		}
	}, mr.WithContext(ctx), mr.WithWorkers(8))
}

func (p *Preheater) IsUnsupported(category string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.unsupported[strings.ToLower(strings.TrimSpace(category))]
	return ok
}

func (p *Preheater) markUnsupported(category string) {
	p.mu.Lock()
	p.unsupported[strings.ToLower(strings.TrimSpace(category))] = struct{}{}
	p.mu.Unlock()
}

func isPackageUnsupported(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "package only supports subscribing")
}

func (p *Preheater) fetchBatchAndCache(ctx context.Context, batch marketDataBatch) error {
	result, err := p.fetchBatch(ctx, batch)
	if err != nil {
		return err
	}
	for _, msg := range batch.msgs {
		data, ok := findBatchData(result, msg.Symbol)
		if !ok {
			logx.Errorf("market batch response missing symbol, topic=%s category=%s market=%s symbol=%s",
				batch.topic, batch.category, batch.market, msg.Symbol)
			continue
		}
		if err := p.cache.Set(ctx, msg, restPayload(batch.topic, data)); err != nil {
			return err
		}
	}
	return nil
}

// FetchQuote verifies one product through iTick REST without publishing it.
// The stale-stream monitor publishes only after checking source freshness.
func (p *Preheater) FetchQuote(ctx context.Context, msg types.ClientMessage) (*types.QuotePayload, error) {
	msg = cache.NormalizeClientMessage(msg)
	msg.Topic = types.TopicQuote
	batch := marketDataBatch{category: msg.CategoryCode, market: msg.Market, topic: types.TopicQuote, msgs: []types.ClientMessage{msg}}
	result, err := p.fetchBatch(ctx, batch)
	if err != nil {
		return nil, err
	}
	data, ok := findBatchData(result, msg.Symbol)
	if !ok {
		return nil, fmt.Errorf("REST quote response missing symbol %s", msg.Symbol)
	}
	payload, ok := restPayload(types.TopicQuote, data).(*types.QuotePayload)
	if !ok || payload == nil {
		return nil, fmt.Errorf("REST quote response has invalid payload for %s", msg.Symbol)
	}
	return payload, nil
}

func (p *Preheater) fetchBatch(ctx context.Context, batch marketDataBatch) (map[string]UpstreamData, error) {
	if p.apiURL == "" || p.restClient == nil || p.cache == nil {
		return nil, fmt.Errorf("REST preheater is not configured")
	}
	base, err := url.Parse(p.apiURL)
	if err != nil {
		return nil, err
	}
	base.Path = path.Join(base.Path, batch.category, string(batch.topic)+"s")
	query := base.Query()
	query.Set("region", batch.market)
	codes := make([]string, 0, len(batch.msgs))
	for _, msg := range batch.msgs {
		codes = append(codes, msg.Symbol)
	}
	query.Set("codes", strings.Join(codes, ","))
	base.RawQuery = query.Encode()

	resp, err := p.restClient.Get(ctx, base.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Code int                     `json:"code"`
		Msg  string                  `json:"msg"`
		Data map[string]UpstreamData `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("REST rejected: code=%d msg=%s", result.Code, result.Msg)
	}
	return result.Data, nil
}

func findBatchData(data map[string]UpstreamData, symbol string) (UpstreamData, bool) {
	if item, ok := data[symbol]; ok {
		return item, true
	}
	for key, item := range data {
		if strings.EqualFold(key, symbol) || strings.EqualFold(item.S, symbol) {
			return item, true
		}
	}
	return UpstreamData{}, false
}

func restPayload(topic types.Topic, data UpstreamData) any {
	switch topic {
	case types.TopicQuote:
		return &types.QuotePayload{LastPrice: data.LD, LastPriceText: data.LDText, Open: data.O, High: data.H, Low: data.L,
			Volume: data.V, Turnover: data.TU, Ts: data.T, Authority: "itick-rest"}
	case types.TopicTick:
		return &types.TickPayload{LastPrice: data.LD, Volume: data.V, Turnover: data.TU, Ts: data.T}
	case types.TopicDepth:
		asks := make([]*types.DepthLevel, 0)
		bids := make([]*types.DepthLevel, 0)
		_ = json.Unmarshal(data.A, &asks)
		_ = json.Unmarshal(data.B, &bids)
		return &types.DepthPayload{Asks: asks, Bids: bids}
	default:
		return nil
	}
}
