package models

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDeliveryPreflightInspectReady(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	input := validDeliveryPreflightInput()
	mock.ExpectQuery("FROM t_trade_symbol AS s").
		WillReturnRows(sqlmock.NewRows([]string{
			"products", "configured", "disabled", "symbol_id", "status",
			"delivery_time", "cutoff", "matching_stop", "server_time",
		}).AddRow(1, 1, 1, 7, 2, 1790323200000, 1790319600000, 1790321400000, 1785405500000))
	mock.ExpectQuery("FROM t_trade_symbol_leverage_config AS l").
		WillReturnRows(sqlmock.NewRows([]string{
			"isolated", "default", "cross", "risk", "valid_risk", "base_risk", "coverage_end",
		}).AddRow(1, 1, 0, 1, 1, 1, 1))
	mock.ExpectQuery("FROM t_itick_price_formula AS f").
		WillReturnRows(sqlmock.NewRows([]string{
			"formulas", "conforming", "fresh", "latest", "server_time",
		}).AddRow(1, 1, 30, 1785405526000, 1785405527000))
	mock.ExpectQuery("FROM t_trade_order AS o").
		WillReturnRows(sqlmock.NewRows([]string{
			"orders", "fills", "positions", "history", "reservations",
			"instructions", "batches", "settlements",
		}).AddRow(0, 0, 0, 0, 0, 0, 0, 0))

	result, err := NewDeliveryPreflightModel(db).Inspect(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if !result.TechnicalReady() {
		t.Fatalf("expected ready result: %+v", result)
	}
	if result.HistoricalFactCount() != 0 {
		t.Fatalf("unexpected historical facts: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeliveryPreflightInspectMissingProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM t_trade_symbol AS s").
		WillReturnRows(sqlmock.NewRows([]string{
			"products", "configured", "disabled", "symbol_id", "status",
			"delivery_time", "cutoff", "matching_stop", "server_time",
		}).AddRow(0, 0, 0, 0, 0, 0, 0, 0, 1785405500000))

	result, err := NewDeliveryPreflightModel(db).Inspect(context.Background(), validDeliveryPreflightInput())
	if err != nil {
		t.Fatal(err)
	}
	if result.TechnicalReady() {
		t.Fatalf("missing product must fail: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeliveryPreflightHistoricalFactBlocksReady(t *testing.T) {
	result := DeliveryPreflightResult{
		ProductCount:                 1,
		ConfiguredProductCount:       1,
		SafeDisabledProductCount:     1,
		SymbolID:                     7,
		ProductStatus:                2,
		DeliveryTimeMs:               300,
		OpenCutoffTimeMs:             200,
		MatchingStopTimeMs:           250,
		ServerTimeMs:                 100,
		IsolatedLeverageConfigCount:  1,
		IsolatedLeverageDefaultCount: 1,
		EnabledRiskTierCount:         1,
		ValidRiskTierCount:           1,
		BaseRiskTierCount:            1,
		RiskCoverageEndCount:         1,
		ConformingFormulaCount:       1,
		FreshSnapshotCount:           1,
		LatestSnapshotTimeMs:         90,
		PriceServerTimeMs:            100,
		SettlementInstructionCount:   1,
	}
	if result.TechnicalReady() {
		t.Fatal("historical delivery fact must block preflight")
	}
}

func TestDeliveryPreflightRejectsInvalidInput(t *testing.T) {
	input := validDeliveryPreflightInput()
	input.FormulaMinInputCount = 2
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := NewDeliveryPreflightModel(db).Inspect(context.Background(), input); err == nil {
		t.Fatal("expected input validation error")
	}
}

func validDeliveryPreflightInput() DeliveryPreflightInput {
	return DeliveryPreflightInput{
		TenantID:               1,
		Symbol:                 "BTCUSDT-20260925",
		SettlementAsset:        "USDT",
		CategoryCode:           "crypto",
		Market:                 "BA",
		PriceSymbol:            "BTCUSDT",
		FormulaVersion:         "delivery-v1",
		FormulaAlgorithm:       2,
		FormulaMaxLookbackMs:   30000,
		FormulaMaxDeviationBps: 200,
		FormulaMinInputCount:   3,
		FormulaIntervalMs:      1000,
	}
}
