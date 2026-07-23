package tasklogic

import (
	"testing"
)

func TestValidateTradeDecimal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value string
		valid bool
	}{
		{"123.456789012345678901", true},
		{"0.000000000000000001", true},
		{"999999999999999999.1", true},
		{"123.4567890123456789012", false},
		{"1000000000000000000", false},
		{"not-a-price", false},
	}
	for _, tt := range tests {
		if got := validateTradeDecimal(tt.value); (got == nil) != tt.valid {
			t.Errorf("validateTradeDecimal(%q) error=%v, valid=%v", tt.value, got, tt.valid)
		}
	}
}

func TestArchiveSnapshotKind(t *testing.T) {
	for input, expected := range map[string]string{"MARK_PRICE": "MARK", "INDEX_PRICE": "INDEX", "FUNDING_RATE": "FUNDING", "DELIVERY_PRICE": "DELIVERY", "SECONDS_SETTLEMENT": "FINAL_QUOTE"} {
		if got := archiveSnapshotKind(input); got != expected {
			t.Fatalf("kind %s: got %s want %s", input, got, expected)
		}
	}
}

func TestFundingSnapshotAllowsSignedRate(t *testing.T) {
	for _, value := range []string{"-0.01", "0", "0.01"} {
		q := &marketQuoteSnapshot{LastPrice: value, QuoteTs: 100, SnapshotID: "s", Confirmed: true}
		if !quoteIsValidAtKind(q, 100, 1000, "FUNDING") {
			t.Fatalf("funding rate %s rejected", value)
		}
	}
}
