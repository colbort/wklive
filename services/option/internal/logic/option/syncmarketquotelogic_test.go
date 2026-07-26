package optionlogic

import (
	"context"
	"strings"
	"testing"

	market "wklive/common/market"
)

func TestSyncAuthoritativeSnapshotValidation(t *testing.T) {
	logic := NewSyncMarketQuoteLogic(context.Background(), nil)
	base := market.AuthoritativeSnapshotEvent{
		Version:         market.AuthoritativeSnapshotEventVersion,
		SnapshotID:      "snapshot-1",
		CategoryCode:    "crypto",
		Market:          "BA",
		Symbol:          "BTCUSDT",
		UnderlyingPrice: "100.25",
	}
	tests := []struct {
		name   string
		mutate func(*market.AuthoritativeSnapshotEvent)
		want   string
	}{
		{name: "version", mutate: func(e *market.AuthoritativeSnapshotEvent) { e.Version++ }, want: "unsupported"},
		{name: "snapshot", mutate: func(e *market.AuthoritativeSnapshotEvent) { e.SnapshotID = "" }, want: "snapshot id"},
		{name: "category", mutate: func(e *market.AuthoritativeSnapshotEvent) { e.CategoryCode = "" }, want: "category"},
		{name: "market", mutate: func(e *market.AuthoritativeSnapshotEvent) { e.Market = "" }, want: "market"},
		{name: "symbol", mutate: func(e *market.AuthoritativeSnapshotEvent) { e.Symbol = "" }, want: "symbol"},
		{name: "price", mutate: func(e *market.AuthoritativeSnapshotEvent) { e.UnderlyingPrice = "invalid" }, want: "price"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := base
			tt.mutate(&event)
			_, err := logic.SyncAuthoritativeSnapshot(event)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestNormalizeQuoteTime(t *testing.T) {
	if got := normalizeQuoteTime(1_700_000_000_123, 1); got != 1_700_000_000 {
		t.Fatalf("millisecond timestamp normalized to %d", got)
	}
	if got := normalizeQuoteTime(0, 123); got != 123 {
		t.Fatalf("fallback timestamp = %d", got)
	}
}
