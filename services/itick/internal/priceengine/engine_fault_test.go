package priceengine

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"wklive/proto/itick"
	"wklive/services/itick/models"
)

type unavailableArchive struct {
	inserted int
}

func (a *unavailableArchive) FindAtOrBefore(context.Context, string, string, string, string, string, int64, int64) (*models.TItickAuthoritativeSnapshot, error) {
	return nil, sql.ErrNoRows
}

func (a *unavailableArchive) InsertImmutableAndEnqueue(context.Context, *models.TItickAuthoritativeSnapshot, string) error {
	a.inserted++
	return nil
}

func TestTemporaryMissingInputDoesNotPublishSnapshot(t *testing.T) {
	archive := &unavailableArchive{}
	engine := &Engine{archive: archive}
	formula := &models.TItickPriceFormula{
		FormulaNo: "BTCUSDT-DELIVERY-v1", SnapshotKind: "DELIVERY",
		CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT",
		Algorithm:     int64(itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN),
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
