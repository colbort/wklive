package models

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestOptionPortfolioLiquidationMetricsExposeDuplicateEvidenceAndCancelStreak(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	conn := sqlx.NewSqlConnFromDB(db)

	mock.ExpectQuery("(?s)liquidation\\.liquidation_scope=2.*liquidation\\.status IN \\(1,2,4,6\\).*HAVING COUNT\\(1\\)>1.*duplicate_wallets").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "count", "oldest"}).
			AddRow(9, 2, 700))
	mock.ExpectQuery("(?s)liquidation_scope=2 AND.*portfolio_maintenance_before<=portfolio_maintenance_after.*portfolio_collateral_after<portfolio_initial_after").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "count", "oldest"}).
			AddRow(9, 1, 710))
	mock.ExpectQuery("(?s)ROW_NUMBER\\(\\) OVER.*sequence_no<=3.*COUNT\\(1\\)=3.*status=7.*cancellation_streaks").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "count", "oldest"}).
			AddRow(9, 1, 720))

	metrics, err := queryOptionPortfolioLiquidationMetricsByTenant(context.Background(), conn)
	if err != nil {
		t.Fatalf("query portfolio liquidation metrics: %v", err)
	}
	assertOperationsMetric(t, metrics, 9, "portfolio_liquidation_duplicate_open", 2, 700)
	assertOperationsMetric(t, metrics, 9, "portfolio_liquidation_evidence_invalid", 1, 710)
	assertOperationsMetric(t, metrics, 9, "portfolio_liquidation_cancel_streak", 1, 720)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestOptionInsuranceTakeoverMetricsExposeRiskWithoutDoubleCountingLots(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	conn := sqlx.NewSqlConnFromDB(db)

	mock.ExpectQuery("(?s)SELECT position\\.tenant_id,COUNT\\(1\\) count,.*MIN\\(lot\\.create_times\\).*EXISTS.*lot\\.trade_id<0.*GROUP BY position\\.tenant_id").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "count", "oldest"}).
			AddRow(9, 2, 700))
	mock.ExpectQuery("(?s)MIN\\(contract\\.expire_time\\).*contract\\.expire_time<=\\?.*EXISTS.*lot\\.trade_id<0").
		WithArgs(int64(87400)).
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "count", "oldest"}).
			AddRow(9, 1, 50000))
	mock.ExpectQuery("(?s)LEFT JOIN t_option_market market.*market\\.id IS NULL.*contract\\.greeks_max_age_seconds<=0").
		WithArgs(int64(970), int64(1000), int64(970), int64(1000), int64(1000), int64(1000)).
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "count", "oldest"}).
			AddRow(9, 1, 800))

	metrics, err := queryOptionInsuranceInventoryMetricsByTenant(
		context.Background(), conn, 1000,
	)
	if err != nil {
		t.Fatalf("query insurance inventory metrics: %v", err)
	}
	assertOperationsMetric(t, metrics, 9, "insurance_takeover_inventory", 2, 700)
	assertOperationsMetric(t, metrics, 9, "insurance_takeover_expiry_due", 1, 50000)
	assertOperationsMetric(t, metrics, 9, "insurance_takeover_market_invalid", 1, 800)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestOptionInsuranceTakeoverAmountsUseUnderlyingAndMarketDimensions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	conn := sqlx.NewSqlConnFromDB(db)

	mock.ExpectQuery("(?s)contract\\.underlying_coin coin.*SUM\\(position\\.position_qty\\*contract\\.multiplier\\).*EXISTS.*lot\\.trade_id<0").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "coin", "amount"}).
			AddRow(9, "BTC", "0.2"))
	mock.ExpectQuery("(?s)contract\\.settle_coin coin.*SUM\\(market\\.mark_price\\*position\\.position_qty\\*contract\\.multiplier\\).*EXISTS.*lot\\.trade_id<0").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "coin", "amount"}).
			AddRow(9, "USDT", "42.5"))
	mock.ExpectQuery("(?s)contract\\.underlying_coin coin.*SUM\\(ABS\\(market\\.delta\\)\\*position\\.position_qty\\*contract\\.multiplier\\).*EXISTS.*lot\\.trade_id<0").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "coin", "amount"}).
			AddRow(9, "BTC", "0.11"))

	amounts, err := queryOptionInsuranceInventoryAmounts(context.Background(), conn)
	if err != nil {
		t.Fatalf("query insurance inventory amounts: %v", err)
	}
	want := []struct {
		category string
		coin     string
		amount   string
	}{
		{"insurance_takeover_underlying_quantity", "BTC", "0.2"},
		{"insurance_takeover_mark_value", "USDT", "42.5"},
		{"insurance_takeover_abs_delta", "BTC", "0.11"},
	}
	if len(amounts) != len(want) {
		t.Fatalf("amounts=%+v want=%+v", amounts, want)
	}
	for index, expected := range want {
		item := amounts[index]
		if item.TenantID != 9 || item.Category != expected.category || item.Coin != expected.coin ||
			!item.Amount.Equal(decimal.RequireFromString(expected.amount)) {
			t.Fatalf("amount[%d]=%+v want=%+v", index, item, expected)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestOptionComboOperationsQueriesStayTenantScoped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	conn := sqlx.NewSqlConnFromDB(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT
  COALESCE(SUM(CASE WHEN status IN (1,5) AND update_times < ? THEN 1 ELSE 0 END),0) stale_count,
  COALESCE(SUM(CASE WHEN status = 8 THEN 1 ELSE 0 END),0) manual_review_count,
  COALESCE(MIN(CASE
    WHEN (status IN (1,5) AND update_times < ?) OR status = 8 THEN update_times
  END),0) oldest
FROM t_option_combo_order
WHERE tenant_id = ?`)).
		WithArgs(int64(940), int64(940), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"stale_count", "manual_review_count", "oldest"}).
			AddRow(2, 1, 901))
	exceptions, err := queryOptionComboExceptions(context.Background(), conn, 9, 940)
	if err != nil {
		t.Fatalf("queryOptionComboExceptions: %v", err)
	}
	if exceptions.StaleCount != 2 || exceptions.ManualReviewCount != 1 || exceptions.Oldest != 901 {
		t.Fatalf("unexpected exceptions: %+v", exceptions)
	}

	mock.ExpectQuery("(?s)SELECT COUNT\\(1\\) count.*WHERE p\\.tenant_id = \\?.*WHERE p\\.id IS NULL AND l\\.tenant_id = \\?.*WHERE o\\.combo_order_id>0 AND l\\.id IS NULL AND o\\.tenant_id = \\?").
		WithArgs(int64(9), int64(9), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	invariantCount, err := queryOptionComboInvariantIssueCount(context.Background(), conn, 9)
	if err != nil {
		t.Fatalf("queryOptionComboInvariantIssueCount: %v", err)
	}
	if invariantCount != 3 {
		t.Fatalf("invariant count=%d want=3", invariantCount)
	}

	mock.ExpectQuery("(?s)SELECT COUNT\\(1\\) count.*FROM t_option_trade.*combo_match_no<>'' AND tenant_id = \\?.*HAVING").
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	incompleteCount, err := queryOptionComboIncompleteMatchCount(context.Background(), conn, 9)
	if err != nil {
		t.Fatalf("queryOptionComboIncompleteMatchCount: %v", err)
	}
	if incompleteCount != 1 {
		t.Fatalf("incomplete match count=%d want=1", incompleteCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
