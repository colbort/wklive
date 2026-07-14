package kline

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"wklive/services/itick/internal/market/cache"
	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type derivedInterval struct {
	name   string
	source string
}

var derivedIntervals = []derivedInterval{
	{name: "5m", source: "1m"},
	{name: "15m", source: "1m"},
	{name: "30m", source: "1m"},
	{name: "1h", source: "1m"},
	{name: "1d", source: "1h"},
	{name: "1w", source: "1d"},
	{name: "1mo", source: "1d"},
}

// DerivedAggregator deterministically rebuilds every supported higher interval
// from persisted lower-interval bars. It is invoked after local 1m writes and
// after REST reconciliation overwrites 1m bars.
type DerivedAggregator struct {
	factory *models.CoinKlineModelFactory
	cache   *cache.MarketDataCache
}

func NewDerivedAggregator(factory *models.CoinKlineModelFactory, marketCache *cache.MarketDataCache) *DerivedAggregator {
	return &DerivedAggregator{factory: factory, cache: marketCache}
}

func (a *DerivedAggregator) Rebuild(minutes []*models.CoinKline) error {
	if a == nil || a.factory == nil || len(minutes) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	type productMinute struct {
		category string
		market   string
		symbol   string
		ts       int64
	}
	items := make(map[productMinute]struct{}, len(minutes))
	for _, item := range minutes {
		if item == nil || item.Ts <= 0 {
			continue
		}
		items[productMinute{strings.ToLower(item.CategoryCode), strings.ToUpper(item.Market), strings.ToUpper(item.Symbol), item.Ts}] = struct{}{}
	}

	var failures []error
	for _, interval := range derivedIntervals {
		buckets := make(map[productMinute]struct{})
		for item := range items {
			start, _ := derivedBucket(item.ts, interval.name)
			item.ts = start
			buckets[item] = struct{}{}
		}
		for bucket := range buckets {
			start, end := derivedBucket(bucket.ts, interval.name)
			if err := a.rebuildBucket(ctx, bucket.category, bucket.market, bucket.symbol, interval.source, interval.name, start, end); err != nil {
				logx.Errorf("rebuild derived kline failed, category=%s market=%s symbol=%s interval=%s ts=%d err=%v",
					bucket.category, bucket.market, bucket.symbol, interval.name, start, err)
				failures = append(failures, err)
			}
		}
	}
	return errors.Join(failures...)
}

func (a *DerivedAggregator) rebuildBucket(ctx context.Context, category, market, symbol, sourceInterval, targetInterval string, start, end int64) error {
	source := a.factory.New(category, sourceInterval)
	target := a.factory.New(category, targetInterval)
	if source == nil || target == nil {
		return nil
	}
	list, err := source.FindRangeByMarketSymbol(ctx, market, symbol, start, end)
	if err != nil || len(list) == 0 {
		return err
	}
	bar := aggregateKlines(category, market, symbol, targetInterval, start, list)
	if err := target.UpsertBySymbolTs(ctx, bar); err != nil {
		return err
	}
	if a.cache != nil {
		msg := types.ClientMessage{Topic: types.TopicKline, CategoryCode: category, Market: market, Symbol: symbol, Interval: targetInterval}
		payload := &types.KlinePayload{Interval: targetInterval, Open: bar.Open, High: bar.High, Low: bar.Low,
			Close: bar.Close, Volume: bar.Volume, Turnover: bar.Turnover, Ts: bar.Ts}
		if err := a.cache.Set(ctx, msg, payload); err != nil {
			logx.Errorf("cache derived kline failed, category=%s market=%s symbol=%s interval=%s err=%v",
				category, market, symbol, targetInterval, err)
		}
	}
	return nil
}

func aggregateKlines(category, market, symbol, interval string, ts int64, list []*models.CoinKline) *models.CoinKline {
	bar := &models.CoinKline{CategoryCode: category, Market: market, Symbol: symbol, Interval: interval, Ts: ts,
		Open: list[0].Open, High: -math.MaxFloat64, Low: math.MaxFloat64, Close: list[len(list)-1].Close}
	for _, item := range list {
		if item.High > bar.High {
			bar.High = item.High
		}
		if item.Low < bar.Low {
			bar.Low = item.Low
		}
		bar.Volume += item.Volume
		bar.Turnover += item.Turnover
	}
	return bar
}

func derivedBucket(ts int64, interval string) (int64, int64) {
	t := time.UnixMilli(ts).UTC()
	var start, end time.Time
	switch interval {
	case "1d":
		start = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 0, 1)
	case "1w":
		offset := (int(t.Weekday()) + 6) % 7
		start = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -offset)
		end = start.AddDate(0, 0, 7)
	case "1mo":
		start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
	default:
		minutes := map[string]int64{"5m": 5, "15m": 15, "30m": 30, "1h": 60}[interval]
		width := minutes * minuteMillis
		startMs := ts / width * width
		return startMs, startMs + width
	}
	return start.UnixMilli(), end.UnixMilli()
}
