package kline

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/internal/pkg/klinewriter"
	"wklive/services/itick/models"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

const minuteMillis int64 = 60_000

type tickBucketKey struct {
	category string
	market   string
	symbol   string
	ts       int64
}

type tickBucket struct {
	openTs, closeTs  int64
	open, high       float64
	low, close       float64
	volume, turnover float64
}

type volumeBaseline struct {
	ts     int64
	volume float64
}

// TickAggregator creates only 1m bars. REST reconciliation remains the source
// of truth and may overwrite these low-latency bars later.
type TickAggregator struct {
	writer    *klinewriter.BatchWriter
	mu        sync.Mutex
	buckets   map[tickBucketKey]*tickBucket
	seen      map[string]int64
	finalized map[string]int64
	baselines map[string]volumeBaseline
	stopOnce  sync.Once
	stopCh    chan struct{}
	doneCh    chan struct{}
	rdb       *redis.Client
	stateTTL  time.Duration
}

type persistedTickState struct {
	Category string  `json:"category"`
	Market   string  `json:"market"`
	Symbol   string  `json:"symbol"`
	Ts       int64   `json:"ts"`
	OpenTs   int64   `json:"openTs"`
	CloseTs  int64   `json:"closeTs"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Volume   float64 `json:"volume"`
	Turnover float64 `json:"turnover"`
}

type persistedBaseline struct {
	ProductKey string  `json:"productKey"`
	Ts         int64   `json:"ts"`
	Volume     float64 `json:"volume"`
}

func NewTickAggregator(writer *klinewriter.BatchWriter, rdb *redis.Client, stateTTL time.Duration) *TickAggregator {
	if stateTTL <= 0 {
		stateTTL = 120 * time.Minute
	}
	return &TickAggregator{writer: writer, buckets: make(map[tickBucketKey]*tickBucket),
		seen: make(map[string]int64), finalized: make(map[string]int64), baselines: make(map[string]volumeBaseline),
		stopCh: make(chan struct{}), doneCh: make(chan struct{}), rdb: rdb, stateTTL: stateTTL}
}

func (a *TickAggregator) Start() {
	a.restore(context.Background())
	go a.run()
}

func (a *TickAggregator) Stop() {
	a.stopOnce.Do(func() { close(a.stopCh) })
	<-a.doneCh
}

func (a *TickAggregator) Add(ctx context.Context, msg types.ClientMessage, tick *types.TickPayload) {
	if tick == nil || tick.Ts <= 0 || tick.LastPrice <= 0 || tick.Volume < 0 || tick.Turnover < 0 ||
		math.IsNaN(tick.LastPrice) || math.IsInf(tick.LastPrice, 0) {
		return
	}
	now := time.Now().UnixMilli()
	if tick.Ts > now+30_000 || tick.Ts < now-10*minuteMillis {
		return
	}
	category := strings.ToLower(strings.TrimSpace(msg.CategoryCode))
	market := strings.ToUpper(strings.TrimSpace(msg.Market))
	symbol := strings.ToUpper(strings.TrimSpace(msg.Symbol))
	if category == "" || market == "" || symbol == "" {
		return
	}
	bucketTs := tick.Ts / minuteMillis * minuteMillis
	productKey := category + ":" + market + ":" + symbol
	fingerprint := fmt.Sprintf("%s:%d:%g:%g:%g", productKey, tick.Ts, tick.LastPrice, tick.Volume, tick.Turnover)

	a.mu.Lock()
	defer a.mu.Unlock()
	if expires, ok := a.seen[fingerprint]; ok && expires > now {
		return
	}
	a.seen[fingerprint] = now + 2*minuteMillis
	if bucketTs <= a.finalized[productKey] {
		return
	}
	volume, turnover := tick.Volume, tick.Turnover
	if category == "stock" {
		volume, turnover = a.stockVolumeDelta(productKey, tick)
	}
	key := tickBucketKey{category: category, market: market, symbol: symbol, ts: bucketTs}
	bucket := a.buckets[key]
	if bucket == nil {
		a.buckets[key] = &tickBucket{openTs: tick.Ts, closeTs: tick.Ts, open: tick.LastPrice,
			high: tick.LastPrice, low: tick.LastPrice, close: tick.LastPrice, volume: volume, turnover: turnover}
		a.persist(ctx, key, a.buckets[key], productKey)
		return
	}
	if tick.Ts < bucket.openTs {
		bucket.openTs, bucket.open = tick.Ts, tick.LastPrice
	}
	if tick.Ts >= bucket.closeTs {
		bucket.closeTs, bucket.close = tick.Ts, tick.LastPrice
	}
	bucket.high = math.Max(bucket.high, tick.LastPrice)
	bucket.low = math.Min(bucket.low, tick.LastPrice)
	bucket.volume += volume
	bucket.turnover += turnover
	a.persist(ctx, key, bucket, productKey)
}

// iTick stock tick volume is a daily cumulative value. The first or reset
// sample establishes a baseline; only positive forward deltas belong to the
// current minute. Stock tick has no reliable turnover field, so use the
// incremental volume multiplied by the latest price until REST correction.
func (a *TickAggregator) stockVolumeDelta(productKey string, tick *types.TickPayload) (float64, float64) {
	previous, ok := a.baselines[productKey]
	a.baselines[productKey] = volumeBaseline{ts: tick.Ts, volume: tick.Volume}
	if !ok || tick.Ts <= previous.ts || tick.Volume < previous.volume {
		if ok && tick.Ts <= previous.ts {
			a.baselines[productKey] = previous
		}
		return 0, 0
	}
	delta := tick.Volume - previous.volume
	return delta, delta * tick.LastPrice
}

func (a *TickAggregator) run() {
	defer close(a.doneCh)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			a.flushClosed(time.Now().UnixMilli())
		case <-a.stopCh:
			a.flushClosed(time.Now().UnixMilli())
			return
		}
	}
}

func (a *TickAggregator) flushClosed(now int64) {
	a.mu.Lock()
	for key, bucket := range a.buckets {
		if key.ts+minuteMillis > now {
			continue
		}
		err := a.writer.Enqueue(&models.CoinKline{CategoryCode: key.category, Market: key.market,
			Symbol: key.symbol, Interval: "1m", Ts: key.ts, Open: bucket.open, High: bucket.high,
			Low: bucket.low, Close: bucket.close, Volume: bucket.volume, Turnover: bucket.turnover,
			Source: models.KlineSourceRealtime, Revision: now, IsClosed: true, Confirmed: false, ActualCount: 1, ExpectedCount: 1})
		if err != nil {
			logx.Errorf("enqueue tick kline failed, will retry, category=%s market=%s symbol=%s ts=%d err=%v",
				key.category, key.market, key.symbol, key.ts, err)
			continue
		}
		delete(a.buckets, key)
		if a.rdb != nil {
			_ = a.rdb.Del(context.Background(), a.bucketRedisKey(key)).Err()
		}
		productKey := key.category + ":" + key.market + ":" + key.symbol
		if key.ts > a.finalized[productKey] {
			a.finalized[productKey] = key.ts
		}
	}
	for fingerprint, expires := range a.seen {
		if expires <= now {
			delete(a.seen, fingerprint)
		}
	}
	a.mu.Unlock()
}

func (a *TickAggregator) persist(ctx context.Context, key tickBucketKey, bucket *tickBucket, productKey string) {
	if a.rdb == nil || bucket == nil {
		return
	}
	state, _ := json.Marshal(persistedTickState{Category: key.category, Market: key.market, Symbol: key.symbol, Ts: key.ts,
		OpenTs: bucket.openTs, CloseTs: bucket.closeTs, Open: bucket.open, High: bucket.high, Low: bucket.low, Close: bucket.close, Volume: bucket.volume, Turnover: bucket.turnover})
	baseline := a.baselines[productKey]
	base, _ := json.Marshal(persistedBaseline{ProductKey: productKey, Ts: baseline.ts, Volume: baseline.volume})
	pipe := a.rdb.Pipeline()
	pipe.Set(ctx, a.bucketRedisKey(key), state, a.stateTTL)
	pipe.Set(ctx, "itick:v1:kline:baseline:"+productKey, base, a.stateTTL)
	_, _ = pipe.Exec(ctx)
}

func (a *TickAggregator) bucketRedisKey(key tickBucketKey) string {
	return fmt.Sprintf("itick:v1:kline:building:%s:%s:%s:%d", key.category, key.market, key.symbol, key.ts)
}

func (a *TickAggregator) restore(ctx context.Context) {
	if a.rdb == nil {
		return
	}
	for _, pattern := range []string{"itick:v1:kline:building:*", "itick:v1:kline:baseline:*"} {
		var cursor uint64
		for {
			keys, next, err := a.rdb.Scan(ctx, cursor, pattern, 200).Result()
			if err != nil {
				break
			}
			for _, key := range keys {
				raw, err := a.rdb.Get(ctx, key).Bytes()
				if err != nil {
					continue
				}
				if strings.Contains(key, ":building:") {
					var v persistedTickState
					if json.Unmarshal(raw, &v) == nil {
						k := tickBucketKey{v.Category, v.Market, v.Symbol, v.Ts}
						a.buckets[k] = &tickBucket{v.OpenTs, v.CloseTs, v.Open, v.High, v.Low, v.Close, v.Volume, v.Turnover}
					}
				} else {
					var v persistedBaseline
					if json.Unmarshal(raw, &v) == nil {
						a.baselines[v.ProductKey] = volumeBaseline{v.Ts, v.Volume}
					}
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
}
