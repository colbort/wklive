package priceengine

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"testing"
	"wklive/proto/itick"
	"wklive/services/itick/models"
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

func TestDeviationAuditPreservesCompleteInputSet(t *testing.T) {
	all := []Input{
		{Price: decimal.NewFromInt(100), Weight: decimal.NewFromInt(1), SnapshotID: "a"},
		{Price: decimal.NewFromInt(101), Weight: decimal.NewFromInt(1), SnapshotID: "b"},
		{Price: decimal.NewFromInt(150), Weight: decimal.NewFromInt(1), SnapshotID: "outlier"},
	}
	accepted, rejected := filterDeviationWithAudit(append([]Input(nil), all...), 200)
	if len(accepted) != 2 || len(rejected) != 1 || rejected[0].SnapshotID != "outlier" {
		t.Fatalf("accepted=%+v rejected=%+v", accepted, rejected)
	}
	if len(all) != 3 || all[2].SnapshotID != "outlier" {
		t.Fatalf("original input set was mutated: %+v", all)
	}
}

func TestDeliverySnapshotIDIsDeterministicFromAuditFact(t *testing.T) {
	formula := &models.TItickPriceFormula{
		Authority:      "price-engine",
		SnapshotKind:   "DELIVERY",
		CategoryCode:   "crypto",
		Market:         "BA",
		Symbol:         "BTCUSDT",
		FormulaNo:      "BTCUSDT-DELIVERY-v1",
		FormulaVersion: "v1",
		Algorithm:      int64(itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN),
	}
	audit := EvaluationAudit{
		FormulaNo:       formula.FormulaNo,
		FormulaVersion:  formula.FormulaVersion,
		Algorithm:       formula.Algorithm,
		TargetTime:      1785217333000,
		AllInputs:       []Input{{Price: decimal.NewFromInt(100), Weight: decimal.NewFromInt(1), SnapshotID: "a"}},
		AcceptedInputs:  []Input{{Price: decimal.NewFromInt(100), Weight: decimal.NewFromInt(1), SnapshotID: "a"}},
		MaxDeviationBps: 200,
	}
	raw, err := json.Marshal(audit)
	if err != nil {
		t.Fatal(err)
	}
	first := deterministicSnapshotID(formula, decimal.NewFromInt(100), raw)
	second := deterministicSnapshotID(formula, decimal.NewFromInt(100), raw)
	if first == "" || first != second {
		t.Fatalf("snapshot id is not deterministic: %q != %q", first, second)
	}
	audit.TargetTime++
	changedRaw, _ := json.Marshal(audit)
	if first == deterministicSnapshotID(formula, decimal.NewFromInt(100), changedRaw) {
		t.Fatal("different target time must produce a different immutable snapshot id")
	}
}

func TestPremiumRateCanBeNegative(t *testing.T) {
	p, err := Calculate(itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE, []Input{{Price: decimal.NewFromInt(99)}, {Price: decimal.NewFromInt(100)}})
	if err != nil || !p.Equal(decimal.RequireFromString("-0.01")) {
		t.Fatalf("premium=%s err=%v", p, err)
	}
}

func TestIndexBasisMarkAppliesSymmetricCap(t *testing.T) {
	tests := []struct {
		name        string
		index       string
		perpetual   string
		capBps      int64
		wantPrice   string
		wantRaw     string
		wantApplied string
	}{
		{name: "inside cap", index: "100", perpetual: "100.5", capBps: 100, wantPrice: "100.5", wantRaw: "0.005", wantApplied: "0.005"},
		{name: "positive basis capped", index: "100", perpetual: "110", capBps: 200, wantPrice: "102", wantRaw: "0.1", wantApplied: "0.02"},
		{name: "negative basis capped", index: "100", perpetual: "90", capBps: 200, wantPrice: "98", wantRaw: "-0.1", wantApplied: "-0.02"},
		{name: "zero basis", index: "100", perpetual: "100", capBps: 200, wantPrice: "100", wantRaw: "0", wantApplied: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, raw, applied, err := CalculateIndexBasis([]Input{
				{Price: decimal.RequireFromString(tt.index)},
				{Price: decimal.RequireFromString(tt.perpetual)},
			}, tt.capBps)
			if err != nil {
				t.Fatal(err)
			}
			if !price.Equal(decimal.RequireFromString(tt.wantPrice)) ||
				!raw.Equal(decimal.RequireFromString(tt.wantRaw)) ||
				!applied.Equal(decimal.RequireFromString(tt.wantApplied)) {
				t.Fatalf("price/raw/applied=%s/%s/%s want=%s/%s/%s",
					price, raw, applied, tt.wantPrice, tt.wantRaw, tt.wantApplied)
			}
		})
	}
}

func TestIndexBasisRejectsMissingRiskBound(t *testing.T) {
	_, _, _, err := CalculateIndexBasis([]Input{
		{Price: decimal.NewFromInt(100)},
		{Price: decimal.NewFromInt(101)},
	}, 0)
	if err == nil {
		t.Fatal("INDEX_BASIS without max basis bound was accepted")
	}
}

func TestIndexBasisSmoothsWithPreviousMark(t *testing.T) {
	price, raw, applied, err := CalculateIndexBasis([]Input{
		{Price: decimal.NewFromInt(100), Weight: decimal.NewFromInt(1)},
		{Price: decimal.NewFromInt(110), Weight: decimal.NewFromInt(1)},
		{Price: decimal.NewFromInt(100), Weight: decimal.NewFromInt(4)},
	}, 200)
	if err != nil {
		t.Fatal(err)
	}
	// Unsmoothed bounded mark is 102. With current:previous weights 1:4,
	// the published mark advances to 100.4.
	if !price.Equal(decimal.RequireFromString("100.4")) ||
		!raw.Equal(decimal.RequireFromString("0.1")) ||
		!applied.Equal(decimal.RequireFromString("0.02")) {
		t.Fatalf("smoothed price/raw/applied=%s/%s/%s", price, raw, applied)
	}
}

func TestDeliveryRequiresThreeAcceptedInputsAfterDeviation(t *testing.T) {
	formula := &models.TItickPriceFormula{SnapshotKind: "DELIVERY", MinInputCount: 1}
	if got := effectiveMinInputCount(formula); got != 3 {
		t.Fatalf("effective delivery minimum=%d, want 3", got)
	}
	formula.MinInputCount = 4
	if got := effectiveMinInputCount(formula); got != 4 {
		t.Fatalf("configured stricter minimum=%d, want 4", got)
	}
}

func TestDuplicateSnapshotIDsDoNotCountAsIndependentInputs(t *testing.T) {
	accepted, rejected := deduplicateInputsWithAudit([]Input{
		{SnapshotID: "same", Price: decimal.NewFromInt(100)},
		{SnapshotID: "same", Price: decimal.NewFromInt(100)},
		{SnapshotID: "other", Price: decimal.NewFromInt(101)},
	})
	if len(accepted) != 2 || len(rejected) != 1 ||
		accepted[0].SnapshotID != "same" || accepted[1].SnapshotID != "other" {
		t.Fatalf("accepted=%+v rejected=%+v", accepted, rejected)
	}
}

func TestReplayEvaluationAuditRejectsTamperedOutput(t *testing.T) {
	audit := EvaluationAudit{
		FormulaNo: "BTCUSDT-INDEX-v2", FormulaVersion: "v2",
		Algorithm:  int64(itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN),
		TargetTime: 1785217333000,
		AllInputs: []Input{
			{Price: decimal.NewFromInt(99), Weight: decimal.NewFromInt(1), SnapshotID: "a"},
			{Price: decimal.NewFromInt(100), Weight: decimal.NewFromInt(1), SnapshotID: "b"},
			{Price: decimal.NewFromInt(101), Weight: decimal.NewFromInt(1), SnapshotID: "c"},
		},
		MinInputCount: 3,
		OutputPrice:   "100",
	}
	audit.AcceptedInputs = append([]Input(nil), audit.AllInputs...)
	raw, _ := json.Marshal(audit)
	price, err := ReplayEvaluationAudit(raw)
	if err != nil || !price.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("valid replay price=%s err=%v", price, err)
	}
	audit.OutputPrice = "100.1"
	raw, _ = json.Marshal(audit)
	if _, err = ReplayEvaluationAudit(raw); err == nil {
		t.Fatal("tampered recorded output passed deterministic replay")
	}
	audit.OutputPrice = "100"
	audit.AcceptedInputs[2].SnapshotID = "b"
	raw, _ = json.Marshal(audit)
	if _, err = ReplayEvaluationAudit(raw); err == nil {
		t.Fatal("duplicate accepted snapshot ids passed replay")
	}
}
