package tasklogic

import (
	"math"
	"testing"

	"wklive/proto/common"
	"wklive/services/liquidity/models"
)

func TestBuildQuoteOrders(t *testing.T) {
	config := &models.TLiquiditySymbolConfig{
		Id: 7, SymbolId: 6, InternalProviderId: 3,
		BaseSpreadBps: 5, MaxSpreadBps: 50, MaxPriceDeviationBps: 40,
		PriceTick: 0.1, QtyStep: 0.001, MinQuoteQty: 0.01,
		MaxQuoteQty: 2, MaxQuoteNotional: 150, QuoteTtlMs: 5_000,
	}
	levels := []*models.TLiquidityStrategyLevel{{
		LevelNo: 1, BidSpreadBps: 5, AskSpreadBps: 10,
		BidQty: 3, AskQty: 1.2345, Enabled: int64(common.Switch_SWITCH_ON),
	}}
	orders := buildQuoteOrders(config, levels, 100, 1_000)
	if len(orders) != 2 {
		t.Fatalf("expected two-sided quotes, got %d", len(orders))
	}
	if orders[0].Side != int64(common.Side_SIDE_BUY) || orders[0].Price != 99.9 || math.Abs(orders[0].Qty-1.501) > 1e-9 {
		t.Fatalf("unexpected bid: %+v", orders[0])
	}
	if orders[1].Side != int64(common.Side_SIDE_SELL) || orders[1].Price != 100.2 || math.Abs(orders[1].Qty-1.234) > 1e-9 {
		t.Fatalf("unexpected ask: %+v", orders[1])
	}
	if orders[0].ExpireAt != 6_000 || orders[1].ExpireAt != 6_000 {
		t.Fatalf("unexpected expiry")
	}
}

func TestParseReferenceSource(t *testing.T) {
	category, market, symbol := parseReferenceSource("crypto:BA:BTCUSDT", "fallback")
	if category != "crypto" || market != "BA" || symbol != "BTCUSDT" {
		t.Fatalf("unexpected parsed source: %s/%s/%s", category, market, symbol)
	}
	_, _, symbol = parseReferenceSource("crypto:BA", "BTCUSDT")
	if symbol != "BTCUSDT" {
		t.Fatalf("expected fallback symbol, got %s", symbol)
	}
}
