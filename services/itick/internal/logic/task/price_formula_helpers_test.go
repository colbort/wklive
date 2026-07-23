package tasklogic

import (
	"testing"

	"wklive/proto/itick"
)

func TestNormalizePriceFormulaReq(t *testing.T) {
	req := &itick.CreatePriceFormulaReq{FormulaNo: "mark-v2", FormulaVersion: "v2", Authority: " PRICE-ENGINE ", SnapshotKind: "mark", CategoryCode: "Crypto", Market: "ba", Symbol: "btcusdt", Algorithm: itick.PriceAlgorithm_PRICE_ALGORITHM_WEIGHTED_MEAN, MaxLookbackMs: 30000, MaxDeviationBps: 100, IntervalMs: 1000, Components: []*itick.PriceFormulaComponent{{Authority: " ITICK-WS ", SnapshotKind: "final_quote", CategoryCode: "Crypto", Market: "ba", Symbol: "btcusdt", Weight: "1.25"}}}
	components, err := normalizePriceFormulaReq(req)
	if err != nil {
		t.Fatal(err)
	}
	if req.Authority != "price-engine" || req.SnapshotKind != "MARK" || req.Symbol != "BTCUSDT" || components[0].Authority != "itick-ws" || components[0].Kind != "FINAL_QUOTE" {
		t.Fatalf("formula was not normalized: req=%+v component=%+v", req, components[0])
	}
}

func TestNormalizePriceFormulaRejectsInvalidPremium(t *testing.T) {
	req := &itick.CreatePriceFormulaReq{FormulaNo: "funding-v1", FormulaVersion: "v1", Authority: "price-engine", SnapshotKind: "FUNDING", Symbol: "BTCUSDT", Algorithm: itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE, MaxLookbackMs: 30000, IntervalMs: 1000, Components: []*itick.PriceFormulaComponent{{Authority: "price-engine", SnapshotKind: "MARK", Symbol: "BTCUSDT", Weight: "1"}}}
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("expected PREMIUM_RATE component count validation error")
	}
}

func TestNormalizePriceFormulaRejectsNonPositiveWeight(t *testing.T) {
	req := &itick.CreatePriceFormulaReq{FormulaNo: "index-v1", FormulaVersion: "v1", Authority: "price-engine", SnapshotKind: "INDEX", Symbol: "BTCUSDT", Algorithm: itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN, MaxLookbackMs: 30000, IntervalMs: 1000, Components: []*itick.PriceFormulaComponent{{Authority: "itick-ws", Symbol: "BTCUSDT", Weight: "0"}}}
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("expected positive weight validation error")
	}
}
