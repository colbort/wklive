package priceengine

import (
	"encoding/json"
	"strings"
	"testing"

	"wklive/proto/itick"

	"github.com/shopspring/decimal"
)

func TestReplayEvaluationAuditWindowReportsContiguousFormulaGroups(t *testing.T) {
	records := [][]byte{
		replayAuditJSON(t, "BTCUSDT-INDEX-v2", "v2", 2_000, "100", "a"),
		replayAuditJSON(t, "BTCUSDT-MARK-v2", "v2", 1_000, "101", "b"),
		replayAuditJSON(t, "BTCUSDT-INDEX-v2", "v2", 1_000, "99", "c"),
		replayAuditJSON(t, "BTCUSDT-MARK-v2", "v2", 2_000, "102", "d"),
	}
	report, err := ReplayEvaluationAuditWindow(records, 1_000)
	if err != nil {
		t.Fatal(err)
	}
	if report.RecordCount != 4 || report.FormulaCount != 2 ||
		report.FirstTargetTime != 1_000 || report.LastTargetTime != 2_000 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if len(report.Formulas) != 2 ||
		report.Formulas[0].FormulaNo != "BTCUSDT-INDEX-v2" ||
		report.Formulas[0].MinimumOutputPrice != "99" ||
		report.Formulas[0].MaximumOutputPrice != "100" {
		t.Fatalf("unexpected formula report: %+v", report.Formulas)
	}
}

func TestReplayEvaluationAuditWindowRejectsGapAndDuplicate(t *testing.T) {
	gap := [][]byte{
		replayAuditJSON(t, "BTCUSDT-INDEX-v2", "v2", 1_000, "99", "a"),
		replayAuditJSON(t, "BTCUSDT-INDEX-v2", "v2", 3_000, "100", "b"),
	}
	if _, err := ReplayEvaluationAuditWindow(gap, 1_000); err == nil ||
		!strings.Contains(err.Error(), "target sequence gap") {
		t.Fatalf("gap error=%v", err)
	}
	duplicate := [][]byte{
		replayAuditJSON(t, "BTCUSDT-INDEX-v2", "v2", 1_000, "99", "a"),
		replayAuditJSON(t, "BTCUSDT-INDEX-v2", "v2", 1_000, "99", "a"),
	}
	if _, err := ReplayEvaluationAuditWindow(duplicate, 0); err == nil ||
		!strings.Contains(err.Error(), "duplicate target_time") {
		t.Fatalf("duplicate error=%v", err)
	}
}

func TestReplayEvaluationAuditRejectsTamperedInputPartition(t *testing.T) {
	inputs := []Input{
		{Price: decimal.NewFromInt(99), Weight: decimal.NewFromInt(1), SnapshotID: "a"},
		{Price: decimal.NewFromInt(100), Weight: decimal.NewFromInt(1), SnapshotID: "b"},
		{Price: decimal.NewFromInt(150), Weight: decimal.NewFromInt(1), SnapshotID: "c"},
	}
	audit := EvaluationAudit{
		FormulaNo: "BTCUSDT-DELIVERY-v2", FormulaVersion: "v2",
		Algorithm:       int64(itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN),
		TargetTime:      1_000,
		AllInputs:       inputs,
		AcceptedInputs:  inputs[:2],
		RejectedInputs:  inputs[2:],
		MaxDeviationBps: 200,
		MinInputCount:   2,
		OutputPrice:     "99.5",
	}
	raw, _ := json.Marshal(audit)
	if _, err := ReplayEvaluationAudit(raw); err != nil {
		t.Fatalf("valid filtered audit rejected: %v", err)
	}
	audit.RejectedInputs = nil
	raw, _ = json.Marshal(audit)
	if _, err := ReplayEvaluationAudit(raw); err == nil ||
		!strings.Contains(err.Error(), "partition mismatch") {
		t.Fatalf("tampered partition error=%v", err)
	}
	audit.RejectedInputs = inputs[2:]
	audit.MinInputCount = 3
	raw, _ = json.Marshal(audit)
	if _, err := ReplayEvaluationAudit(raw); err == nil ||
		!strings.Contains(err.Error(), "below minimum") {
		t.Fatalf("minimum input error=%v", err)
	}
}

func replayAuditJSON(
	t *testing.T, formulaNo, version string, target int64, price, snapshotID string,
) []byte {
	t.Helper()
	input := Input{
		Price:  decimal.RequireFromString(price),
		Weight: decimal.NewFromInt(1), SnapshotID: snapshotID,
	}
	audit := EvaluationAudit{
		FormulaNo: formulaNo, FormulaVersion: version,
		Algorithm:  int64(itick.PriceAlgorithm_PRICE_ALGORITHM_WEIGHTED_MEAN),
		TargetTime: target, AllInputs: []Input{input},
		AcceptedInputs: []Input{input}, MinInputCount: 1, OutputPrice: price,
	}
	raw, err := json.Marshal(audit)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
