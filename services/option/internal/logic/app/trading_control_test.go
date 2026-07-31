package applogic

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestOptionOrderPriceBandIncludesBoundaries(t *testing.T) {
	mark := decimal.NewFromInt(100)
	ratio := decimal.RequireFromString("0.10")
	for _, test := range []struct {
		price string
		ok    bool
	}{
		{price: "89.99", ok: false},
		{price: "90", ok: true},
		{price: "100", ok: true},
		{price: "110", ok: true},
		{price: "110.01", ok: false},
	} {
		lower, upper, ok := optionOrderPriceBand(
			decimal.RequireFromString(test.price), mark, ratio,
		)
		if ok != test.ok {
			t.Fatalf("price=%s band=[%s,%s] ok=%t want=%t", test.price, lower, upper, ok, test.ok)
		}
	}
}

func TestOptionOrderPriceBandRejectsMissingControl(t *testing.T) {
	if _, _, ok := optionOrderPriceBand(
		decimal.NewFromInt(100), decimal.NewFromInt(100), decimal.Zero,
	); ok {
		t.Fatal("zero price band must mean not configured, not unlimited")
	}
}

func TestOptionExposureLimitExceeded(t *testing.T) {
	limit := decimal.NewFromInt(10)
	if optionExposureLimitExceeded(decimal.NewFromInt(7), decimal.NewFromInt(3), limit) {
		t.Fatal("exact limit boundary should be admitted")
	}
	if !optionExposureLimitExceeded(decimal.NewFromInt(7), decimal.RequireFromString("3.01"), limit) {
		t.Fatal("exposure above limit must be rejected")
	}
	if !optionExposureLimitExceeded(decimal.Zero, decimal.NewFromInt(1), decimal.Zero) {
		t.Fatal("zero limit must be treated as unconfigured and rejected")
	}
}
