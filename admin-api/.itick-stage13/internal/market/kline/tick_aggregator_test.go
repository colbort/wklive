package kline

import (
	"context"
	"testing"
	"time"

	"wklive/services/itick/internal/market/types"
)

func TestTickAggregatorUsesEventTimeAndDeduplicates(t *testing.T) {
	a := NewTickAggregator(nil)
	now := time.Now().UnixMilli()
	bucketTs := now / minuteMillis * minuteMillis
	msg := types.ClientMessage{CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT"}
	a.Add(context.Background(), msg, &types.TickPayload{LastPrice: 11, Volume: 2, Turnover: 22, Ts: bucketTs + 20_000})
	a.Add(context.Background(), msg, &types.TickPayload{LastPrice: 10, Volume: 1, Turnover: 10, Ts: bucketTs + 10_000})
	a.Add(context.Background(), msg, &types.TickPayload{LastPrice: 12, Volume: 3, Turnover: 36, Ts: bucketTs + 30_000})
	a.Add(context.Background(), msg, &types.TickPayload{LastPrice: 12, Volume: 3, Turnover: 36, Ts: bucketTs + 30_000})

	key := tickBucketKey{category: "crypto", market: "BA", symbol: "BTCUSDT", ts: bucketTs}
	bucket := a.buckets[key]
	if bucket == nil {
		t.Fatal("expected bucket")
	}
	if bucket.open != 10 || bucket.high != 12 || bucket.low != 10 || bucket.close != 12 {
		t.Fatalf("unexpected OHLC: %+v", bucket)
	}
	if bucket.volume != 6 || bucket.turnover != 68 {
		t.Fatalf("duplicate tick was not ignored: %+v", bucket)
	}
}

func TestTickAggregatorRejectsFinalizedBucket(t *testing.T) {
	a := NewTickAggregator(nil)
	now := time.Now().UnixMilli()
	bucketTs := now / minuteMillis * minuteMillis
	productKey := "crypto:BA:BTCUSDT"
	a.finalized[productKey] = bucketTs
	a.Add(context.Background(), types.ClientMessage{CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT"},
		&types.TickPayload{LastPrice: 10, Volume: 1, Ts: bucketTs + 1})
	if len(a.buckets) != 0 {
		t.Fatal("finalized bucket must not be recreated by late ticks")
	}
}
