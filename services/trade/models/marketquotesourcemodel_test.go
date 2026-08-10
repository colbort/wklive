package models

import "testing"

func TestMarketQuoteSourceSource(t *testing.T) {
	tests := []struct {
		name   string
		value  *MarketQuoteSource
		expect string
	}{
		{name: "normalized forex source", value: &MarketQuoteSource{CategoryCode: " Forex ", Market: "gb", Symbol: "usdcny"}, expect: "forex:GB:USDCNY"},
		{name: "missing market", value: &MarketQuoteSource{CategoryCode: "forex", Symbol: "USDCNY"}, expect: ""},
		{name: "nil", value: nil, expect: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.value.Source(); got != test.expect {
				t.Fatalf("Source() = %q, want %q", got, test.expect)
			}
		})
	}
}
