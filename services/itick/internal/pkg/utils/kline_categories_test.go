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
