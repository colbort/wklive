package models

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestContractReadinessModelInspectDetailed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	input := validReadinessInput()
	mock.ExpectQuery(`(?s)FROM t_market_authority_registry`).
		WithArgs(
			"source-a", "source-b", "source-c",
			"source-a", "source-b", "source-c",
			"source-a", "source-b", "source-c",
		).
		WillReturnRows(sqlmock.NewRows([]string{"sources", "providers", "public_rest", "price_engine"}).AddRow(3, 3, 3, 1))
	mock.ExpectQuery(`(?s)FROM t_market_price_formula AS f`).
		WithArgs(
			2, "index-v1", int64(30000), int64(200), 3, int64(1000), 3,
			"source-a", "market-a", "1",
			"source-b", "market-b", "1",
			"source-c", "market-c", "1",
			3,
			"mark-v2", int64(30000), int64(200), int64(1000),
			"1", "perpetual-source", "PERPETUAL", "1", "4",
			"funding-v1", int64(30000), int64(1000),
			2, "delivery-v1", int64(30000), int64(200),
			3, int64(1000), 3,
			"source-a", "market-a", "1",
			"source-b", "market-b", "1",
			"source-c", "market-c", "1",
			3,
			"crypto", "BA", "BTCUSDT",
		).
		WillReturnRows(sqlmock.NewRows([]string{"index", "mark", "funding", "delivery"}).
			AddRow(1, 1, 1, 1))
	mock.ExpectQuery(`(?s)JOIN t_market_authoritative_snapshot AS s`).
		WithArgs(
			"crypto", "BA", "BTCUSDT",
			"source-a", "source-b", "source-c",
			"crypto", "BA", "BTCUSDT", int64(30000),
		).
		WillReturnRows(sqlmock.NewRows([]string{"sources", "outputs"}).AddRow(3, 4))
	mock.ExpectQuery(`(?s)FROM t_asset_platform_account`).
		WithArgs(int64(900101), "USDT").
		WillReturnRows(sqlmock.NewRows([]string{"insurance", "fee"}).AddRow(1, 1))
	mock.ExpectQuery(`(?s)settlement_window_seconds\*1000=\?.*settlement_price_algorithm=\?.*FROM t_trade_symbol AS s`).
		WithArgs(
			"BTCUSDT-PERP",
			"BTCUSDT-20260925",
			int64(30000),
			"delivery-v1",
			int64(900101),
			"USDT",
		).
		WillReturnRows(sqlmock.NewRows([]string{"perpetual", "delivery"}).AddRow(1, 1))
	mock.ExpectQuery(`(?s)FROM t_contract_insurance_fund_account`).
		WithArgs(int64(900101), "USDT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`(?s)FROM t_market_snapshot_outbox`).
		WillReturnRows(sqlmock.NewRows([]string{"pending", "processing", "failed", "manual", "oldest", "server_now", "reconciliation", "settlement"}).
			AddRow(0, 0, 0, 0, 0, 1_000_000, 0, 0))

	result, err := NewContractReadinessModel(db).Inspect(context.Background(), input, true)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Detailed ||
		result.ActiveSourceAuthorityCount != 3 ||
		result.DistinctSourceProviderCount != 3 ||
		result.PublicRestSourceCount != 3 ||
		result.PriceEngineAuthorityCount != 1 ||
		result.IndexFormulaCount != 1 ||
		result.MarkFormulaCount != 1 ||
		result.FundingFormulaCount != 1 ||
		result.DeliveryFormulaCount != 1 ||
		result.FreshSourceCount != 3 ||
		result.FreshOutputKindCount != 4 ||
		result.InsuranceFundCount != 1 ||
		result.FeeRevenueCount != 1 ||
		result.PerpetualContractCount != 1 ||
		result.DeliveryContractCount != 1 ||
		result.InsuranceConfigCount != 1 ||
		result.OpenOutboxCount != 0 ||
		result.UnhealthyOutboxCount != 0 ||
		result.OpenReconciliationCount != 0 ||
		result.OpenSettlementCount != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestContractReadinessModelInspectOpenOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)FROM t_market_snapshot_outbox`).
		WillReturnRows(sqlmock.NewRows([]string{"pending", "processing", "failed", "manual", "oldest", "server_now", "reconciliation", "settlement"}).
			AddRow(1, 1, 0, 0, 900_000, 910_000, 3, 4))
	result, err := NewContractReadinessModel(db).Inspect(
		context.Background(),
		ContractReadinessInput{},
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Detailed ||
		result.OpenOutboxCount != 2 ||
		result.UnhealthyOutboxCount != 0 ||
		result.OpenReconciliationCount != 3 ||
		result.OpenSettlementCount != 4 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestContractReadinessModelInspectRejectsStaleOrFailedOutbox(t *testing.T) {
	tests := []struct {
		name string
		row  *sqlmock.Rows
		want int64
	}{
		{
			name: "stale processing",
			row: sqlmock.NewRows([]string{"pending", "processing", "failed", "manual", "oldest", "server_now", "reconciliation", "settlement"}).
				AddRow(0, 2, 0, 0, 900_000, 960_001, 0, 0),
			want: 2,
		},
		{
			name: "failed row",
			row: sqlmock.NewRows([]string{"pending", "processing", "failed", "manual", "oldest", "server_now", "reconciliation", "settlement"}).
				AddRow(0, 0, 1, 0, 959_999, 960_000, 0, 0),
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery(`(?s)FROM t_market_snapshot_outbox`).WillReturnRows(tt.row)
			result, err := NewContractReadinessModel(db).Inspect(context.Background(), ContractReadinessInput{}, false)
			if err != nil {
				t.Fatal(err)
			}
			if result.UnhealthyOutboxCount != tt.want {
				t.Fatalf("unhealthy outbox=%d want=%d result=%+v", result.UnhealthyOutboxCount, tt.want, result)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestContractReadinessModelRejectsInvalidDetailedInput(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	input := validReadinessInput()
	input.SourceAuthorities = []string{"source-a", "source-b"}
	_, err = NewContractReadinessModel(db).Inspect(context.Background(), input, true)
	if err == nil {
		t.Fatal("expected source validation error")
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestContractReadinessModelRejectsSubSecondDeliveryWindow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	input := validReadinessInput()
	input.DeliveryMaxLookbackMs = 30500
	_, err = NewContractReadinessModel(db).Inspect(context.Background(), input, true)
	if err == nil {
		t.Fatal("expected delivery window validation error")
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestContractReadinessModelRejectsInvalidSourceWeights(t *testing.T) {
	tests := []struct {
		name    string
		weights []string
	}{
		{name: "missing", weights: []string{"1", "1"}},
		{name: "zero", weights: []string{"1", "0", "1"}},
		{name: "invalid", weights: []string{"1", "bad", "1"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			input := validReadinessInput()
			input.DeliverySourceWeights = test.weights
			_, err = NewContractReadinessModel(db).Inspect(context.Background(), input, true)
			if err == nil {
				t.Fatal("expected source-weight validation error")
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPlaceholdersAndStringArgs(t *testing.T) {
	if got := placeholders(3); got != "?,?,?" {
		t.Fatalf("placeholders(3) = %q", got)
	}
	args := appendStringArgs([]any{"prefix"}, []string{"a", "b"})
	if len(args) != 3 || args[0] != "prefix" || args[1] != "a" || args[2] != "b" {
		t.Fatalf("unexpected args: %#v", args)
	}
	if got := sourceIdentityWeightPredicate(2); got !=
		"(j.authority=? AND j.market=? AND j.category_code=f.category_code AND j.symbol=f.symbol "+
			"AND j.weight=CAST(? AS DECIMAL(36,18))) OR "+
			"(j.authority=? AND j.market=? AND j.category_code=f.category_code AND j.symbol=f.symbol "+
			"AND j.weight=CAST(? AS DECIMAL(36,18)))" {
		t.Fatalf("unexpected source-weight predicate: %q", got)
	}
	args = appendSourceIdentityWeightArgs(
		nil,
		[]string{"a", "b"},
		[]string{"market-a", "market-b"},
		[]string{"1", "2"},
	)
	if len(args) != 6 ||
		args[0] != "a" || args[1] != "market-a" || args[2] != "1" ||
		args[3] != "b" || args[4] != "market-b" || args[5] != "2" {
		t.Fatalf("unexpected source-weight args: %#v", args)
	}
}

func validReadinessInput() ContractReadinessInput {
	return ContractReadinessInput{
		SourceAuthorities:       []string{"source-a", "source-b", "source-c"},
		SourceMarkets:           []string{"market-a", "market-b", "market-c"},
		IndexSourceWeights:      []string{"1", "1", "1"},
		DeliverySourceWeights:   []string{"1", "1", "1"},
		CategoryCode:            "crypto",
		Market:                  "BA",
		PriceSymbol:             "BTCUSDT",
		PerpetualSymbol:         "BTCUSDT-PERP",
		DeliverySymbol:          "BTCUSDT-20260925",
		PerpetualPriceAuthority: "perpetual-source",
		PerpetualPriceMarket:    "PERPETUAL",
		TenantID:                900101,
		SettlementCoin:          "USDT",
		IndexAlgorithm:          2,
		IndexFormulaVersion:     "index-v1",
		IndexMaxDeviationBps:    200,
		MarkFormulaVersion:      "mark-v2",
		MarkMaxBasisBps:         200,
		MarkCurrentWeight:       "1",
		MarkPreviousWeight:      "4",
		FundingFormulaVersion:   "funding-v1",
		PriceFormulaIntervalMs:  1000,
		DeliveryAlgorithm:       2,
		DeliveryFormulaVersion:  "delivery-v1",
		DeliveryMaxLookbackMs:   30000,
		DeliveryMaxDeviationBps: 200,
	}
}
