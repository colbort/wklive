package priceengine

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"wklive/proto/market"
	"wklive/services/market/models"

	"github.com/shopspring/decimal"
)

type unavailableArchive struct {
	inserted int
}

func (a *unavailableArchive) FindAtOrBefore(context.Context, string, string, string, string, string, int64, int64) (*models.TMarketAuthoritativeSnapshot, error) {
	return nil, sql.ErrNoRows
}

func (a *unavailableArchive) InsertImmutableAndEnqueue(context.Context, *models.TMarketAuthoritativeSnapshot, string) error {
	a.inserted++
	return nil
}

func TestTemporaryMissingInputDoesNotPublishSnapshot(t *testing.T) {
	archive := &unavailableArchive{}
	engine := &Engine{archive: archive}
	formula := &models.TMarketPriceFormula{
		FormulaNo: "BTCUSDT-DELIVERY-v1", SnapshotKind: "DELIVERY",
		CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT",
		Algorithm:     int64(market.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN),
		Components:    `[{"authority":"itick-ws","kind":"FINAL_QUOTE","category_code":"crypto","market":"BA","symbol":"BTCUSDT","weight":"1"}]`,
		MaxLookbackMs: 30_000,
	}
	err := engine.evaluate(context.Background(), formula, 1785217333000)
	if !errors.Is(err, ErrInputUnavailable) {
		t.Fatalf("error=%v, want ErrInputUnavailable", err)
	}
	if archive.inserted != 0 {
		t.Fatalf("missing input published %d snapshots", archive.inserted)
	}
	if !strings.Contains(err.Error(), "formula=BTCUSDT-DELIVERY-v1") ||
		!strings.Contains(err.Error(), "authority=itick-ws") {
		t.Fatalf("missing input error is not diagnosable: %v", err)
	}
}

type pricedArchive struct {
	prices      map[string]string
	sourceTimes map[string]int64
	lookups     map[string]int64
	inserted    int
	last        *models.TMarketAuthoritativeSnapshot
}

func (a *pricedArchive) FindAtOrBefore(_ context.Context, _, _, _, market, _ string, target, _ int64) (*models.TMarketAuthoritativeSnapshot, error) {
	if a.lookups == nil {
		a.lookups = make(map[string]int64)
	}
	a.lookups[market] = target
	price, ok := a.prices[market]
	if !ok {
		return nil, sql.ErrNoRows
	}
	sourceTimestamp := target
	if configured, exists := a.sourceTimes[market]; exists {
		sourceTimestamp = configured
	}
	return &models.TMarketAuthoritativeSnapshot{
		SnapshotId:      "snapshot-" + market,
		Price:           decimal.RequireFromString(price),
		SourceTimestamp: sourceTimestamp,
	}, nil
}

func (a *pricedArchive) InsertImmutableAndEnqueue(_ context.Context, snapshot *models.TMarketAuthoritativeSnapshot, _ string) error {
	a.inserted++
	a.last = snapshot
	return nil
}

func TestDeliveryDoesNotPublishWhenDeviationLeavesFewerThanThreeInputs(t *testing.T) {
	archive := &pricedArchive{prices: map[string]string{
		"SOURCE_A": "100",
		"SOURCE_B": "101",
		"SOURCE_C": "150",
	}}
	engine := &Engine{archive: archive}
	formula := &models.TMarketPriceFormula{
		FormulaNo: "BTCUSDT-DELIVERY-v1", FormulaVersion: "v1",
		Authority: "price-engine", SnapshotKind: "DELIVERY",
		CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT",
		Algorithm: int64(market.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN),
		Components: `[
			{"authority":"itick-ws","kind":"FINAL_QUOTE","category_code":"crypto","market":"SOURCE_A","symbol":"BTCUSDT","weight":"1"},
			{"authority":"itick-ws","kind":"FINAL_QUOTE","category_code":"crypto","market":"SOURCE_B","symbol":"BTCUSDT","weight":"1"},
			{"authority":"itick-ws","kind":"FINAL_QUOTE","category_code":"crypto","market":"SOURCE_C","symbol":"BTCUSDT","weight":"1"}
		]`,
		MaxLookbackMs: 30_000, MaxDeviationBps: 200, MinInputCount: 3,
	}
	err := engine.evaluate(context.Background(), formula, 1785217333000)
	if !errors.Is(err, ErrInputUnavailable) {
		t.Fatalf("error=%v, want ErrInputUnavailable", err)
	}
	if !strings.Contains(err.Error(), "accepted=2 required=3 rejected=1") {
		t.Fatalf("unexpected input quorum error: %v", err)
	}
	if archive.inserted != 0 {
		t.Fatalf("insufficient input quorum published %d snapshots", archive.inserted)
	}
}

func TestIndexBasisPublishesBoundedAuditableMark(t *testing.T) {
	archive := &pricedArchive{prices: map[string]string{
		"INDEX":     "100",
		"PERPETUAL": "110",
	}}
	engine := &Engine{archive: archive}
	formula := &models.TMarketPriceFormula{
		FormulaNo: "BTCUSDT-MARK-v2", FormulaVersion: "v2",
		Authority: "price-engine", SnapshotKind: "MARK",
		CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT",
		Algorithm: int64(market.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS),
		Components: `[
			{"authority":"price-engine","kind":"INDEX","category_code":"crypto","market":"INDEX","symbol":"BTCUSDT","weight":"1"},
			{"authority":"itick-ws","kind":"FINAL_QUOTE","category_code":"crypto","market":"PERPETUAL","symbol":"BTCUSDT","weight":"1"}
		]`,
		MaxLookbackMs: 30_000, MaxDeviationBps: 200, MinInputCount: 2,
	}
	if err := engine.evaluate(context.Background(), formula, 1785217333000); err != nil {
		t.Fatal(err)
	}
	if archive.inserted != 1 || archive.last == nil {
		t.Fatalf("published snapshots=%d last=%+v", archive.inserted, archive.last)
	}
	if !archive.last.Price.Equal(decimal.NewFromInt(102)) {
		t.Fatalf("bounded MARK=%s want=102", archive.last.Price)
	}
	var audit EvaluationAudit
	if err := json.Unmarshal([]byte(archive.last.RawPayload), &audit); err != nil {
		t.Fatal(err)
	}
	if audit.RawBasisRate != "0.1" || audit.AppliedBasisRate != "0.02" {
		t.Fatalf("basis audit raw/applied=%s/%s", audit.RawBasisRate, audit.AppliedBasisRate)
	}
	replayed, err := ReplayEvaluationAudit([]byte(archive.last.RawPayload))
	if err != nil || !replayed.Equal(decimal.NewFromInt(102)) {
		t.Fatalf("replayed MARK=%s err=%v", replayed, err)
	}
}

func TestIndexBasisSmoothingReadsOnlyPreviousMark(t *testing.T) {
	const target = int64(1785217333000)
	archive := &pricedArchive{prices: map[string]string{
		"INDEX": "100", "PERPETUAL": "110", "PREVIOUS": "100",
	}, sourceTimes: map[string]int64{
		"INDEX": target - 100, "PERPETUAL": target - 50, "PREVIOUS": target - 20_000,
	}}
	engine := &Engine{archive: archive}
	formula := &models.TMarketPriceFormula{
		FormulaNo: "BTCUSDT-MARK-v3", FormulaVersion: "v3",
		Authority: "price-engine", SnapshotKind: "MARK",
		CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT",
		Algorithm: int64(market.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS),
		Components: `[
			{"authority":"price-engine","kind":"INDEX","category_code":"crypto","market":"INDEX","symbol":"BTCUSDT","weight":"1"},
			{"authority":"itick-ws","kind":"FINAL_QUOTE","category_code":"crypto","market":"PERPETUAL","symbol":"BTCUSDT","weight":"1"},
			{"authority":"price-engine","kind":"MARK","category_code":"crypto","market":"PREVIOUS","symbol":"BTCUSDT","weight":"4"}
		]`,
		MaxLookbackMs: 30_000, MaxDeviationBps: 200, MinInputCount: 3,
	}
	if err := engine.evaluate(context.Background(), formula, target); err != nil {
		t.Fatal(err)
	}
	if archive.lookups["PREVIOUS"] != target-1 {
		t.Fatalf("previous MARK lookup target=%d want=%d", archive.lookups["PREVIOUS"], target-1)
	}
	if archive.last == nil || !archive.last.Price.Equal(decimal.RequireFromString("100.4")) {
		t.Fatalf("smoothed MARK=%v want=100.4", archive.last)
	}
	if archive.last.SourceTimestamp != target-100 {
		t.Fatalf(
			"smoothed MARK source time=%d want current market minimum=%d",
			archive.last.SourceTimestamp,
			target-100,
		)
	}
	if replayed, err := ReplayEvaluationAudit([]byte(archive.last.RawPayload)); err != nil ||
		!replayed.Equal(decimal.RequireFromString("100.4")) {
		t.Fatalf("replayed smoothed MARK=%s err=%v", replayed, err)
	}
}
