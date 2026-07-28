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
