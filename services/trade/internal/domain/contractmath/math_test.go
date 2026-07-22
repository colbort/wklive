package contractmath

import (
	"testing"

	"github.com/shopspring/decimal"

	"wklive/proto/trade"
)

func TestCalculateTradeValues(t *testing.T) {
	linear, err := CalculateTradeValues(int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR), decimal.NewFromInt(2), decimal.RequireFromString("0.001"), decimal.NewFromInt(50000))
	if err != nil {
		t.Fatal(err)
	}
	if !linear.QuoteNotional.Equal(decimal.NewFromInt(100)) || !linear.SettlementNotional.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("linear values = %+v, want quote/settlement 100", linear)
	}
	if fee := CalculateFee(linear, decimal.RequireFromString("0.0005")); !fee.Equal(decimal.RequireFromString("0.05")) {
		t.Fatalf("linear fee = %s, want 0.05", fee)
	}

	inverse, err := CalculateTradeValues(int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE), decimal.NewFromInt(100), decimal.NewFromInt(100), decimal.NewFromInt(50000))
	if err != nil {
		t.Fatal(err)
	}
	if !inverse.QuoteNotional.Equal(decimal.NewFromInt(10000)) || !inverse.SettlementNotional.Equal(decimal.RequireFromString("0.2")) {
		t.Fatalf("inverse values = %+v, want quote 10000/base 0.2", inverse)
	}
	if fee := CalculateFee(inverse, decimal.RequireFromString("0.0005")); !fee.Equal(decimal.RequireFromString("0.0001")) {
		t.Fatalf("inverse fee = %s, want 0.0001", fee)
	}
	if margin, err := CalculateMargin(inverse, 10); err != nil || !margin.Equal(decimal.RequireFromString("0.02")) {
		t.Fatalf("inverse margin = %s err=%v, want 0.02", margin, err)
	}
}
