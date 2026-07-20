package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestSettlementSnapshotDigestBindsRevisionAndFormula(t *testing.T) {
	base := &SettlementSnapshot{Kind: "FUNDING", MarkPrice: "100", IndexPrice: "99", FundingRate: "0.01", SourceTimestamp: 1000, SnapshotTimestamp: 1001, Revision: 7, FormulaVersion: "premium-v1", Confirmed: true}
	a := snapshotDigest(base)
	copy := *base
	copy.Revision++
	if a == snapshotDigest(&copy) {
		t.Fatal("revision must change snapshot id")
	}
	copy = *base
	copy.FormulaVersion = "premium-v2"
	if a == snapshotDigest(&copy) {
		t.Fatal("formula version must change snapshot id")
	}
	copy = *base
	copy.SnapshotTimestamp++
	if a != snapshotDigest(&copy) {
		t.Fatal("read time must not change source snapshot id")
	}
}

func TestAuthoritativeQuoteArchivePreservesDecimalAndHistoricalTime(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	cache := NewMarketDataCache(client)
	ctx := context.Background()
	msg := ClientMessage{Topic: TopicQuote, CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT"}
	for _, quote := range []*QuotePayload{
		{LastPrice: 1.2345678901234567, LastPriceText: "1.234567890123456789", Ts: 1000, Authority: "itick-ws"},
		{LastPrice: 2, LastPriceText: "2.000000000000000001", Ts: 2000, Authority: "itick-ws"},
	} {
		if _, err := cache.PublishAuthoritativeQuote(ctx, msg, quote); err != nil {
			t.Fatal(err)
		}
	}
	got, err := cache.FindAuthoritativeQuoteAt(ctx, msg, "itick-ws", 1500, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Price != "1.234567890123456789" || got.SourceTimestamp != 1000 || !got.Confirmed {
		t.Fatalf("unexpected historical snapshot: %#v", got)
	}
}
