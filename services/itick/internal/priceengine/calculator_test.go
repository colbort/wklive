package priceengine

import (
	"github.com/shopspring/decimal"
	"testing"
	"wklive/proto/itick"
)

func TestCalculate(t *testing.T) {
	p, e := Calculate(itick.PriceAlgorithm_PRICE_ALGORITHM_WEIGHTED_MEAN, []Input{{Price: decimal.NewFromInt(100), Weight: decimal.NewFromInt(1)}, {Price: decimal.NewFromInt(110), Weight: decimal.NewFromInt(3)}})
	if e != nil || !p.Equal(decimal.RequireFromString("107.5")) {
		t.Fatalf("weighted=%s err=%v", p, e)
	}
	p, e = Calculate(itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN, []Input{{Price: decimal.NewFromInt(3)}, {Price: decimal.NewFromInt(1)}, {Price: decimal.NewFromInt(2)}})
	if e != nil || !p.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("median=%s err=%v", p, e)
	}
}

func TestPremiumRateCanBeNegative(t *testing.T) {
	p, err := Calculate(itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE, []Input{{Price: decimal.NewFromInt(99)}, {Price: decimal.NewFromInt(100)}})
	if err != nil || !p.Equal(decimal.RequireFromString("-0.01")) {
		t.Fatalf("premium=%s err=%v", p, err)
	}
}
