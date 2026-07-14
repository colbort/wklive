package kline

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/internal/pkg/klinewriter"
	"wklive/services/itick/models"

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
}

func NewTickAggregator(writer *klinewriter.BatchWriter) *TickAggregator {
	return &TickAggregator{writer: writer, buckets: make(map[tickBucketKey]*tickBucket),
		seen: make(map[string]int64), finalized: make(map[string]int64), baselines: make(map[string]volumeBaseline),
		stopCh: make(chan struct{}), doneCh: make(chan struct{})}
}

func (a *TickAggregator) Start() { go a.run() }

func (a *TickAggregator) Stop() {
	a.stopOnce.Do(func() { close(a.stopCh) })
	<-a.doneCh
}

func (a *TickAggregator) Add(_ context.Context, msg types.ClientMessage, tick *types.TickPayload) {
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
			Low: bucket.low, Close: bucket.close, Volume: bucket.volume, Turnover: bucket.turnover})
		if err != nil {
			logx.Errorf("enqueue tick kline failed, will retry, category=%s market=%s symbol=%s ts=%d err=%v",
				key.category, key.market, key.symbol, key.ts, err)
			continue
		}
		delete(a.buckets, key)
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
