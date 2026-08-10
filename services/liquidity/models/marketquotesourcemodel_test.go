package models

import "testing"

func TestMarketQuoteSourceSource(t *testing.T) {
	source := (&MarketQuoteSource{CategoryCode: " Forex ", Market: "gb", Symbol: "usdcny"}).Source()
	if source != "forex:GB:USDCNY" {
		t.Fatalf("Source() = %q, want %q", source, "forex:GB:USDCNY")
	}
}
