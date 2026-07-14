package kline

import (
	"testing"
	"time"

	"wklive/services/itick/models"
)

func TestAggregateKlines(t *testing.T) {
	list := []*models.CoinKline{
		{Open: 10, High: 12, Low: 9, Close: 11, Volume: 2, Turnover: 21},
		{Open: 11, High: 14, Low: 10, Close: 13, Volume: 3, Turnover: 38},
	}
	bar := aggregateKlines("crypto", "BA", "BTCUSDT", "5m", 1000, list)
	if bar.Open != 10 || bar.High != 14 || bar.Low != 9 || bar.Close != 13 {
		t.Fatalf("unexpected OHLC: %+v", bar)
	}
	if bar.Volume != 5 || bar.Turnover != 59 {
		t.Fatalf("unexpected volume: %+v", bar)
	}
}

func TestDerivedBucketCalendarBoundaries(t *testing.T) {
	ts := time.Date(2026, 7, 14, 12, 3, 0, 0, time.UTC).UnixMilli()
	start, end := derivedBucket(ts, "5m")
	if start != time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC).UnixMilli() || end-start != 5*minuteMillis {
		t.Fatalf("unexpected 5m bucket: %d - %d", start, end)
	}
	start, end = derivedBucket(ts, "1w")
	if time.UnixMilli(start).UTC().Weekday() != time.Monday || end-start != int64(7*24*time.Hour/time.Millisecond) {
		t.Fatalf("unexpected weekly bucket: %d - %d", start, end)
	}
	start, end = derivedBucket(ts, "1mo")
	if time.UnixMilli(start).UTC().Day() != 1 || time.UnixMilli(end).UTC().Month() != time.August {
		t.Fatalf("unexpected monthly bucket: %d - %d", start, end)
	}
}
