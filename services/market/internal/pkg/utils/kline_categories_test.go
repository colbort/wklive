package utils

import "testing"

func TestStockHolidayMarketsByCode(t *testing.T) {
	got := StockHolidayMarketsByCode()
	china := got["CN"]
	if len(china) != 2 || china[0] != "SH" || china[1] != "SZ" {
		t.Fatalf("unexpected CN markets: %v", china)
	}
	if us := got["US"]; len(us) != 1 || us[0] != "US" {
		t.Fatalf("unexpected US markets: %v", us)
	}
}

func TestStockExchangeMatchesMarket(t *testing.T) {
	tests := []struct {
		market   string
		exchange string
		want     bool
	}{
		{market: "SH", exchange: "SSE", want: true},
		{market: "SZ", exchange: "SZSE", want: true},
		{market: "US", exchange: "NASDAQ", want: true},
		{market: "us", exchange: "nyse", want: true},
		{market: "SH", exchange: "SZSE", want: false},
		{market: "SZ", exchange: "SSE", want: false},
		{market: "US", exchange: "", want: false},
		{market: "CN", exchange: "SSE", want: false},
	}

	for _, test := range tests {
		if got := StockExchangeMatchesMarket(test.market, test.exchange); got != test.want {
			t.Errorf("StockExchangeMatchesMarket(%q, %q)=%v want=%v", test.market, test.exchange, got, test.want)
		}
	}
}
