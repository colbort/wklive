package utils

import (
	"fmt"
	"strings"
	"time"
	"wklive/proto/itick"
)

type KlineIntervalMeta struct {
	Name   string
	KType  itick.KlineType
	Stream string
}

var KlineIntervals = []KlineIntervalMeta{
	{Name: "1m", KType: itick.KlineType_KLINE_TYPE_1M, Stream: "kline@1"},
	{Name: "5m", KType: itick.KlineType_KLINE_TYPE_5M, Stream: "kline@2"},
	{Name: "15m", KType: itick.KlineType_KLINE_TYPE_15M, Stream: "kline@3"},
	{Name: "30m", KType: itick.KlineType_KLINE_TYPE_30M, Stream: "kline@4"},
	{Name: "1h", KType: itick.KlineType_KLINE_TYPE_60M, Stream: "kline@5"},
	{Name: "1d", KType: itick.KlineType_KLINE_TYPE_1D, Stream: "kline@8"},
	{Name: "1w", KType: itick.KlineType_KLINE_TYPE_1W, Stream: "kline@9"},
	{Name: "1mo", KType: itick.KlineType_KLINE_TYPE_1MO, Stream: "kline@10"},
}

var DefaultKlineIntervals = func() []string {
	out := make([]string, 0, len(KlineIntervals))
	for _, v := range KlineIntervals {
		out = append(out, v.Name)
	}
	return out
}()

var DefaultKTypes = func() []int32 {
	out := make([]int32, 0, len(KlineIntervals))
	for _, v := range KlineIntervals {
		out = append(out, int32(v.KType))
	}
	return out
}()

var kTypeToName map[itick.KlineType]string
var nameToStream map[string]string
var streamToName map[string]string
var nameAliasToName map[string]string

func init() {
	kTypeToName = make(map[itick.KlineType]string, len(KlineIntervals))
	nameToStream = make(map[string]string, len(KlineIntervals))
	streamToName = make(map[string]string, len(KlineIntervals))
	nameAliasToName = make(map[string]string, len(KlineIntervals)+1)

	for _, v := range KlineIntervals {
		lowerName := strings.ToLower(strings.TrimSpace(v.Name))
		lowerStream := strings.ToLower(strings.TrimSpace(v.Stream))

		kTypeToName[v.KType] = v.Name
		nameToStream[lowerName] = v.Stream
		streamToName[lowerStream] = v.Name
		nameAliasToName[lowerName] = v.Name
	}

	// 别名
	nameAliasToName["60m"] = "1h"
}

func KlineTypeToInterval(kType itick.KlineType) string {
	if v, ok := kTypeToName[kType]; ok {
		return v
	}
	return "1m"
}

func IntervalToStream(interval string) (string, error) {
	key := strings.ToLower(strings.TrimSpace(interval))
	if canonical, ok := nameAliasToName[key]; ok {
		if stream, ok := nameToStream[strings.ToLower(canonical)]; ok {
			return stream, nil
		}
	}
	return "", fmt.Errorf("unsupported interval: %s", interval)
}

func StreamToInterval(stream string) (string, bool) {
	v, ok := streamToName[strings.ToLower(strings.TrimSpace(stream))]
	return v, ok
}

func NormalizeMarket(market string) string {
	return strings.ToUpper(strings.TrimSpace(market))
}

func NormalizeSymbol(symbol string) string {
	return strings.TrimSpace(symbol)
}

func KTypeToIntervalName(kType int32) string {
	return kTypeToName[itick.KlineType(kType)]
}

func IntervalMillis(interval string) int64 {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "1m":
		return int64(time.Minute / time.Millisecond)
	case "5m":
		return int64(5 * time.Minute / time.Millisecond)
	case "15m":
		return int64(15 * time.Minute / time.Millisecond)
	case "30m":
		return int64(30 * time.Minute / time.Millisecond)
	case "1h", "60m":
		return int64(time.Hour / time.Millisecond)
	case "1d":
		return int64(24 * time.Hour / time.Millisecond)
	case "1w":
		return int64(7 * 24 * time.Hour / time.Millisecond)
	case "1mo":
		return int64(30 * 24 * time.Hour / time.Millisecond)
	default:
		return 0
	}
}

func LastClosedTs(nowMs int64, interval string) int64 {
	if nowMs <= 0 {
		nowMs = time.Now().UnixMilli()
	}

	name := strings.ToLower(strings.TrimSpace(interval))
	t := time.UnixMilli(nowMs).UTC()

	switch name {
	case "1mo":
		firstOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		return firstOfMonth.AddDate(0, -1, 0).UnixMilli()
	case "1w":
		dayOffset := (int(t.Weekday()) + 6) % 7
		weekStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -dayOffset)
		return weekStart.AddDate(0, 0, -7).UnixMilli()
	}

	intervalMs := IntervalMillis(name)
	if intervalMs <= 0 {
		return 0
	}
	return nowMs/intervalMs*intervalMs - intervalMs
}

func RecentCheckBars(interval string) int {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "1m", "5m":
		return 5
	case "15m", "30m", "1h", "60m":
		return 4
	default:
		return 3
	}
}

func IncrementalPages(interval string) int {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "1m":
		return 8
	case "5m":
		return 4
	default:
		return 2
	}
}
