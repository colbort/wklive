package trade

import (
	"testing"

	"wklive/app-api/internal/types"
)

func TestNormalizePlaceOrderReqScalesAmountOnly(t *testing.T) {
	req := &types.PlaceOrderReq{
		Qty:    "1.23",
		Amount: "1.23",
	}

	got := normalizePlaceOrderReq(req)
	if got.Qty != "1.23" {
		t.Fatalf("qty = %q, want original qty", got.Qty)
	}
	if got.Amount != "123" {
		t.Fatalf("amount = %q, want 123", got.Amount)
	}
}

func TestScaleTradeMinorText(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: ""},
		{name: "integer", value: "1", want: "100"},
		{name: "one decimal", value: "1.2", want: "120"},
		{name: "two decimals", value: "1.23", want: "123"},
		{name: "more decimals", value: "0.001", want: "0.1"},
		{name: "positive sign", value: "+2", want: "200"},
		{name: "negative", value: "-1.23", want: "-123"},
		{name: "invalid", value: "abc", want: "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scaleTradeMinorText(tt.value); got != tt.want {
				t.Fatalf("scaleTradeMinorText(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
