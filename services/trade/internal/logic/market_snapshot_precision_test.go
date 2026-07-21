package logic

import "testing"

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
