package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"wklive/services/itick/internal/marketdata/types"

	"github.com/redis/go-redis/v9"
)

type CacheEnvelope struct {
	Topic        string          `json:"topic"`
	CategoryCode string          `json:"categoryCode"`
	Symbol       string          `json:"symbol"`
	Market       string          `json:"market,omitempty"`
	Interval     string          `json:"interval,omitempty"`
	Payload      json.RawMessage `json:"payload"`
}

type MarketDataCache struct {
	rdb          *redis.Client
	mu           sync.RWMutex
	quoteHandler func(context.Context, types.ClientMessage, *types.QuotePayload)
	tickHandler  func(context.Context, types.ClientMessage, *types.TickPayload)
}

type CachedMarketData struct {
	Message types.ClientMessage
	Payload any
	Version string
}

func NewMarketDataCache(rdb *redis.Client) *MarketDataCache {
	return &MarketDataCache{rdb: rdb}
}

func (b *MarketDataCache) Set(ctx context.Context, msg types.ClientMessage, payload any) error {
	msg = NormalizeClientMessage(msg)

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	env := CacheEnvelope{
		Topic:        string(msg.Topic),
		CategoryCode: msg.CategoryCode,
		Symbol:       msg.Symbol,
		Market:       msg.Market,
		Interval:     msg.Interval,
		Payload:      raw,
	}

	bs, err := json.Marshal(env)
	if err != nil {
		return err
	}

	if err := b.rdb.Set(ctx, marketDataKey(msg), bs, marketDataTTL(msg.Topic)).Err(); err != nil {
		return err
	}
	if quote, ok := payload.(*types.QuotePayload); ok && quote != nil {
		b.mu.RLock()
		handler := b.quoteHandler
		b.mu.RUnlock()
		if handler != nil {
			go handler(ctx, msg, quote)
		}
	}
	if tick, ok := payload.(*types.TickPayload); ok && tick != nil {
		b.mu.RLock()
		handler := b.tickHandler
		b.mu.RUnlock()
		if handler != nil {
			handler(ctx, msg, tick)
		}
	}
	return nil
}

func (b *MarketDataCache) SetTickHandler(handler func(context.Context, types.ClientMessage, *types.TickPayload)) {
	b.mu.Lock()
	b.tickHandler = handler
	b.mu.Unlock()
}

func (b *MarketDataCache) SetQuoteHandler(handler func(context.Context, types.ClientMessage, *types.QuotePayload)) {
	b.mu.Lock()
	b.quoteHandler = handler
	b.mu.Unlock()
}

func (b *MarketDataCache) Read(ctx context.Context, msg types.ClientMessage) (any, string, error) {
	raw, err := b.rdb.Get(ctx, marketDataKey(msg)).Result()
	if err != nil {
		return nil, "", err
	}
	var env CacheEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		return nil, "", err
	}
	payload, err := decodeMarketDataPayload(msg.Topic, env.Payload)
	return payload, raw, err
}

func (b *MarketDataCache) ReadMany(ctx context.Context, msgs []types.ClientMessage) ([]CachedMarketData, error) {
	if len(msgs) == 0 {
		return nil, nil
	}
	keys := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		keys = append(keys, marketDataKey(msg))
	}
	values, err := b.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	out := make([]CachedMarketData, 0, len(values))
	for i, value := range values {
		raw, ok := value.(string)
		if !ok || raw == "" {
			continue
		}
		var env CacheEnvelope
		if err := json.Unmarshal([]byte(raw), &env); err != nil {
			continue
		}
		payload, err := decodeMarketDataPayload(msgs[i].Topic, env.Payload)
		if err != nil {
			continue
		}
		out = append(out, CachedMarketData{Message: msgs[i], Payload: payload, Version: raw})
	}
	return out, nil
}

func marketDataKey(msg types.ClientMessage) string {
	msg = NormalizeClientMessage(msg)
	if msg.Topic == types.TopicKline {
		return fmt.Sprintf("itick:v1:kline:%s:%s:%s:%s", msg.CategoryCode, msg.Market, msg.Symbol, msg.Interval)
	}
	return fmt.Sprintf("itick:v1:%s:%s:%s:%s", msg.Topic, msg.CategoryCode, msg.Market, msg.Symbol)
}

func marketDataTTL(topic types.Topic) time.Duration {
	switch topic {
	case types.TopicDepth:
		return 5 * time.Minute
	case types.TopicKline:
		return 24 * time.Hour
	default:
		return 30 * time.Minute
	}
}

func decodeMarketDataPayload(topic types.Topic, raw json.RawMessage) (any, error) {
	switch topic {
	case types.TopicQuote:
		var v types.QuotePayload
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case types.TopicTick:
		var v types.TickPayload
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case types.TopicDepth:
		var v types.DepthPayload
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case types.TopicKline:
		var v types.KlinePayload
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil
	}

	return nil, nil
}
