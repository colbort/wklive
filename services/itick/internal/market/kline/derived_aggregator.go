package kline

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"wklive/services/itick/internal/market/cache"
	marketcalendar "wklive/services/itick/internal/market/calendar"
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
	factory  *models.CoinKlineModelFactory
	cache    *cache.MarketDataCache
	calendar *marketcalendar.Resolver
}

func NewDerivedAggregator(factory *models.CoinKlineModelFactory, marketCache *cache.MarketDataCache, calendar *marketcalendar.Resolver) *DerivedAggregator {
	return &DerivedAggregator{factory: factory, cache: marketCache, calendar: calendar}
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
		buckets := make(map[productMinute]map[int64]struct{})
		for item := range items {
			start, _ := a.bucket(ctx, item.category, item.market, item.ts, interval.name)
			product := productMinute{category: item.category, market: item.market, symbol: item.symbol}
			if buckets[product] == nil {
				buckets[product] = make(map[int64]struct{})
			}
			buckets[product][start] = struct{}{}
		}
		for product, starts := range buckets {
			if err := a.rebuildProductBuckets(ctx, product.category, product.market, product.symbol, interval.source, interval.name, starts); err != nil {
				logx.Errorf("rebuild derived kline failed, category=%s market=%s symbol=%s interval=%s err=%v",
					product.category, product.market, product.symbol, interval.name, err)
				failures = append(failures, err)
			}
		}
	}
	return errors.Join(failures...)
}

func (a *DerivedAggregator) bucket(ctx context.Context, category, market string, ts int64, interval string) (int64, int64) {
	if a.calendar != nil {
		return a.calendar.Bucket(ctx, category, market, "", ts, interval)
	}
	return derivedBucket(ts, interval)
}

func (a *DerivedAggregator) rebuildProductBuckets(ctx context.Context, category, market, symbol, sourceInterval, targetInterval string, starts map[int64]struct{}) error {
	source := a.factory.New(category, sourceInterval)
	target := a.factory.New(category, targetInterval)
	if source == nil || target == nil || len(starts) == 0 {
		return nil
	}
	var minStart, maxEnd int64
	ends := make(map[int64]int64, len(starts))
	for start := range starts {
		_, end := a.bucket(ctx, category, market, start, targetInterval)
		ends[start] = end
		if minStart == 0 || start < minStart {
			minStart = start
		}
		if end > maxEnd {
			maxEnd = end
		}
	}
	list, err := source.FindRangeByMarketSymbol(ctx, market, symbol, minStart, maxEnd)
	if err != nil || len(list) == 0 {
		return err
	}
	grouped := make(map[int64][]*models.CoinKline, len(starts))
	for _, item := range list {
		start, _ := a.bucket(ctx, category, market, item.Ts, targetInterval)
		if _, wanted := starts[start]; wanted {
			grouped[start] = append(grouped[start], item)
		}
	}
	bars := make([]*models.CoinKline, 0, len(grouped))
	for start, sourceBars := range grouped {
		if len(sourceBars) == 0 {
			continue
		}
		bars = append(bars, aggregateKlines(category, market, symbol, sourceInterval, targetInterval, start, ends[start], sourceBars))
	}
	if err := target.BulkUpsertBySymbolTs(ctx, bars); err != nil {
		return err
	}
	if a.cache != nil {
		for _, bar := range bars {
			msg := types.ClientMessage{Topic: types.TopicKline, CategoryCode: category, Market: market, Symbol: symbol, Interval: targetInterval}
			payload := &types.KlinePayload{Interval: targetInterval, Open: bar.Open, High: bar.High, Low: bar.Low,
				Close: bar.Close, Volume: bar.Volume, Turnover: bar.Turnover, Ts: bar.Ts, Source: bar.Source,
				Revision: bar.Revision, IsClosed: bar.IsClosed, Confirmed: bar.Confirmed,
				ActualCount: bar.ActualCount, ExpectedCount: bar.ExpectedCount}
			if err := a.cache.Set(ctx, msg, payload); err != nil {
				logx.Errorf("cache derived kline failed, category=%s market=%s symbol=%s interval=%s err=%v",
					category, market, symbol, targetInterval, err)
			}
		}
	}
	return nil
}

func aggregateKlines(category, market, symbol, sourceInterval, interval string, ts, end int64, list []*models.CoinKline) *models.CoinKline {
	expected := expectedSourceCount(sourceInterval, interval, ts, end)
	bar := &models.CoinKline{CategoryCode: category, Market: market, Symbol: symbol, Interval: interval, Ts: ts,
		Open: list[0].Open, High: -math.MaxFloat64, Low: math.MaxFloat64, Close: list[len(list)-1].Close,
		Source: models.KlineSourceDerived, Revision: time.Now().UnixMilli(), IsClosed: end <= time.Now().UnixMilli(),
		ActualCount: int32(len(list)), ExpectedCount: int32(expected)}
	bar.Confirmed = bar.IsClosed && expected > 0 && len(list) == expected
	for _, item := range list {
		if item.High > bar.High {
			bar.High = item.High
		}
		if item.Low < bar.Low {
			bar.Low = item.Low
		}
		bar.Volume += item.Volume
		bar.Turnover += item.Turnover
		bar.Confirmed = bar.Confirmed && item.Confirmed
	}
	return bar
}

func expectedSourceCount(source, target string, start, end int64) int {
	switch source {
	case "1m":
		return int((end - start) / minuteMillis)
	case "1h":
		return int((end - start) / int64(time.Hour/time.Millisecond))
	case "1d":
		return int(time.UnixMilli(end).UTC().Sub(time.UnixMilli(start).UTC()).Hours() / 24)
	default:
		return 0
	}
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
