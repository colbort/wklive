package tasklogic

import (
	"testing"

	"wklive/proto/market"
)

func TestNormalizePriceFormulaReq(t *testing.T) {
	req := &market.CreatePriceFormulaReq{FormulaNo: "mark-v2", FormulaVersion: "v2", Authority: " PRICE-ENGINE ", SnapshotKind: "mark", CategoryCode: "Crypto", Market: "ba", Symbol: "btcusdt", Algorithm: market.PriceAlgorithm_PRICE_ALGORITHM_WEIGHTED_MEAN, MaxLookbackMs: 30000, MaxDeviationBps: 100, IntervalMs: 1000, Components: []*market.PriceFormulaComponent{{Authority: " ITICK-WS ", SnapshotKind: "final_quote", CategoryCode: "Crypto", Market: "ba", Symbol: "btcusdt", Weight: "1.25"}}}
	components, err := normalizePriceFormulaReq(req)
	if err != nil {
		t.Fatal(err)
	}
	if req.Authority != "price-engine" || req.SnapshotKind != "MARK" || req.Symbol != "BTCUSDT" || components[0].Authority != "itick-ws" || components[0].Kind != "FINAL_QUOTE" {
		t.Fatalf("formula was not normalized: req=%+v component=%+v", req, components[0])
	}
}

func TestNormalizePriceFormulaRejectsInvalidPremium(t *testing.T) {
	req := &market.CreatePriceFormulaReq{FormulaNo: "funding-v1", FormulaVersion: "v1", Authority: "price-engine", SnapshotKind: "FUNDING", Symbol: "BTCUSDT", Algorithm: market.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE, MaxLookbackMs: 30000, IntervalMs: 1000, Components: []*market.PriceFormulaComponent{{Authority: "price-engine", SnapshotKind: "MARK", Symbol: "BTCUSDT", Weight: "1"}}}
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("expected PREMIUM_RATE component count validation error")
	}
}

func TestNormalizePriceFormulaRejectsNonPositiveWeight(t *testing.T) {
	req := &market.CreatePriceFormulaReq{FormulaNo: "index-v1", FormulaVersion: "v1", Authority: "price-engine", SnapshotKind: "INDEX", Symbol: "BTCUSDT", Algorithm: market.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN, MaxLookbackMs: 30000, IntervalMs: 1000, Components: []*market.PriceFormulaComponent{{Authority: "itick-ws", Symbol: "BTCUSDT", Weight: "0"}}}
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("expected positive weight validation error")
	}
}

func TestNormalizeDeliveryFormulaRequiresThreeAcceptedInputs(t *testing.T) {
	components := []*market.PriceFormulaComponent{
		{Authority: "itick-ws", SnapshotKind: "FINAL_QUOTE", Market: "BA", Symbol: "BTCUSDT", Weight: "1"},
		{Authority: "source-b", SnapshotKind: "FINAL_QUOTE", Market: "BB", Symbol: "BTCUSDT", Weight: "1"},
		{Authority: "source-c", SnapshotKind: "FINAL_QUOTE", Market: "BC", Symbol: "BTCUSDT", Weight: "1"},
	}
	req := &market.CreatePriceFormulaReq{
		FormulaNo: "delivery-v1", FormulaVersion: "v1", Authority: "price-engine",
		SnapshotKind: "DELIVERY", CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT",
		Algorithm: market.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN, MaxLookbackMs: 30000,
		IntervalMs: 1000, Components: components, MinInputCount: 2,
	}
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("expected DELIVERY min_input_count below 3 to be rejected")
	}
	req.MinInputCount = 3
	if _, err := normalizePriceFormulaReq(req); err != nil {
		t.Fatalf("expected valid DELIVERY formula: %v", err)
	}
}

func TestNormalizeIndexBasisMarkFormula(t *testing.T) {
	req := &market.CreatePriceFormulaReq{
		FormulaNo: "BTCUSDT-MARK-v2", FormulaVersion: "v2", Authority: "price-engine",
		SnapshotKind: "MARK", CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT",
		Algorithm:     market.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS,
		MaxLookbackMs: 30000, MaxDeviationBps: 200, MinInputCount: 2, IntervalMs: 1000,
		Components: []*market.PriceFormulaComponent{
			{Authority: "price-engine", SnapshotKind: "INDEX", CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT", Weight: "1"},
			{Authority: "itick-ws", SnapshotKind: "FINAL_QUOTE", CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT", Weight: "1"},
		},
	}
	if _, err := normalizePriceFormulaReq(req); err != nil {
		t.Fatalf("valid INDEX_BASIS MARK rejected: %v", err)
	}
	req.MaxDeviationBps = 0
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("INDEX_BASIS without cap accepted")
	}
	req.MaxDeviationBps = 200
	req.Components[0], req.Components[1] = req.Components[1], req.Components[0]
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("INDEX_BASIS with reversed component semantics accepted")
	}

	req.Components[0], req.Components[1] = req.Components[1], req.Components[0]
	req.Components = append(req.Components, &market.PriceFormulaComponent{
		Authority: "price-engine", SnapshotKind: "MARK", CategoryCode: "crypto",
		Market: "BA", Symbol: "BTCUSDT", Weight: "4",
	})
	req.MinInputCount = 3
	if _, err := normalizePriceFormulaReq(req); err != nil {
		t.Fatalf("valid smoothed INDEX_BASIS rejected: %v", err)
	}
	req.Components[2].Authority = "itick-ws"
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("previous MARK from non-output authority accepted")
	}
}

func TestNormalizeIndexFormulaRequiresThreeDistinctAuthorities(t *testing.T) {
	req := &market.CreatePriceFormulaReq{
		FormulaNo: "BTCUSDT-INDEX-v2", FormulaVersion: "v2", Authority: "price-engine",
		SnapshotKind: "INDEX", CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT",
		Algorithm:     market.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN,
		MaxLookbackMs: 30000, MinInputCount: 3, IntervalMs: 1000,
		Components: []*market.PriceFormulaComponent{
			{Authority: "itick-ws", SnapshotKind: "FINAL_QUOTE", CategoryCode: "crypto", Market: "SOURCE_A", Symbol: "BTCUSDT", Weight: "1"},
			{Authority: "source-b", SnapshotKind: "FINAL_QUOTE", CategoryCode: "crypto", Market: "SOURCE_B", Symbol: "BTCUSDT", Weight: "1"},
			{Authority: "source-c", SnapshotKind: "FINAL_QUOTE", CategoryCode: "crypto", Market: "SOURCE_C", Symbol: "BTCUSDT", Weight: "1"},
		},
	}
	if _, err := normalizePriceFormulaReq(req); err != nil {
		t.Fatalf("valid three-source INDEX rejected: %v", err)
	}
	req.Components[2].Authority = "itick-ws"
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("same INDEX authority with a different market accepted")
	}
	req.Components[2].Authority = "source-c"
	req.Components[2] = req.Components[1]
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("duplicate INDEX source accepted")
	}
	req.Components = req.Components[:2]
	req.MinInputCount = 2
	if _, err := normalizePriceFormulaReq(req); err == nil {
		t.Fatal("two-source INDEX accepted")
	}
}
