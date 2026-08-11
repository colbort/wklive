package utils

import "testing"

func TestStockMarketQuoteCurrency(t *testing.T) {
	tests := map[string]string{
		"SH": "CNY",
		"sz": "CNY",
		"HK": "HKD",
		"US": "USD",
		"JP": "JPY",
		"GB": "GBP",
	}

	for market, want := range tests {
		got, ok := StockMarketQuoteCurrency(market)
		if !ok || got != want {
			t.Fatalf("StockMarketQuoteCurrency(%q) = %q, %v; want %q, true", market, got, ok, want)
		}
	}
}

func TestDefaultProductAssets(t *testing.T) {
	base, quote := DefaultProductAssets("stock", "SH", "600930")
	if base != "600930" || quote != "CNY" {
		t.Fatalf("unexpected stock assets: base=%q quote=%q", base, quote)
	}

	base, quote = DefaultProductAssets("crypto", "BA", "BTCUSDT")
	if base != "" || quote != "" {
		t.Fatalf("non-stock assets must not be guessed: base=%q quote=%q", base, quote)
	}
}
