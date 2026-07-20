package logic

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestValidateFundingConservation(t *testing.T) {
	if err := validateFundingConservation(map[string]decimal.Decimal{"USDT": decimal.Zero}); err != nil {
		t.Fatal(err)
	}
	if err := validateFundingConservation(map[string]decimal.Decimal{"USDT": decimal.RequireFromString("0.000000000000000001")}); err == nil {
		t.Fatal("expected unbalanced batch to be rejected")
	}
}
