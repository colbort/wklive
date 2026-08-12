package client

import (
	"testing"
	"time"

	"wklive/services/market/internal/market/types"
)

func TestQuoteNeedsRecovery(t *testing.T) {
	now := time.UnixMilli(10_000)
	tests := []struct {
		name  string
		quote *types.QuotePayload
		want  bool
	}{
		{name: "missing", want: true},
		{name: "missing timestamp", quote: &types.QuotePayload{}, want: true},
		{name: "fresh", quote: &types.QuotePayload{Ts: 9_500}, want: false},
		{name: "boundary", quote: &types.QuotePayload{Ts: 9_000}, want: false},
		{name: "stale", quote: &types.QuotePayload{Ts: 8_999}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := quoteNeedsRecovery(now, test.quote, time.Second); got != test.want {
				t.Fatalf("quoteNeedsRecovery()=%t want=%t", got, test.want)
			}
		})
	}
}

func TestStaleQuoteRecoveryConfigDefaults(t *testing.T) {
	cfg := (StaleQuoteRecoveryConfig{}).withDefaults()
	if cfg.CheckInterval <= 0 || cfg.StaleAfter <= 0 || cfg.StartupGrace <= 0 ||
		cfg.RestMaxAge <= 0 || cfg.Cooldown <= 0 || cfg.BatchSize <= 0 {
		t.Fatalf("invalid defaults: %+v", cfg)
	}
}

func TestRestQuoteCanRecover(t *testing.T) {
	now := time.UnixMilli(10_000)
	if restQuoteCanRecover(now, nil, &types.QuotePayload{Ts: 1_000}, time.Second) {
		t.Fatal("stale REST quote must not recover or publish")
	}
	if restQuoteCanRecover(now, &types.QuotePayload{Ts: 9_500}, &types.QuotePayload{Ts: 9_500}, time.Second) {
		t.Fatal("REST quote with the same timestamp must not trigger resubscribe")
	}
	if !restQuoteCanRecover(now, &types.QuotePayload{Ts: 8_000}, &types.QuotePayload{Ts: 9_500}, time.Second) {
		t.Fatal("fresh newer REST quote should recover the product")
	}
}
