package models

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestQueryOptionOperationsOverviewMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_OPERATIONS_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_OPERATIONS_TEST_DSN is not set")
	}
	now := time.Now().Unix()
	result, err := QueryOptionOperationsOverview(
		context.Background(), sqlx.NewMysql(dsn), 9, now-60, now-60,
	)
	if err != nil {
		t.Fatalf("query overview: %v", err)
	}
	if os.Getenv("OPTION_OPERATIONS_EXPECT_SEEDED") != "1" {
		return
	}
	if result.AssetFailedCount != 1 ||
		result.OpenReconciliationCount != 1 ||
		result.StaleRiskAccountCount != 1 ||
		result.PendingExerciseCount != 1 ||
		result.FailedSettlementCount != 1 ||
		result.ExceptionLiquidationCount != 1 ||
		result.PhysicalExceptionCount != 1 {
		t.Fatalf("unexpected overview counts: %+v", result)
	}
	if result.ComboStaleCount != 0 ||
		result.ComboManualReviewCount != 0 ||
		result.ComboInvariantIssueCount != 0 ||
		result.ComboIncompleteMatchCount != 0 {
		t.Fatalf("unexpected combo overview counts: %+v", result)
	}
	assertCoinAmount(t, result.InsuranceLedger, "USDT", "7")
	assertCoinAmount(t, result.BackstopLiability, "USDT", "3")
	assertCoinAmount(t, result.UnresolvedDeficit, "USDT", "2")
}

func TestQueryOptionComboOperationsOverviewSeededMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_OPERATIONS_TEST_DSN")
	if dsn == "" || os.Getenv("OPTION_COMBO_OPERATIONS_EXPECT_SEEDED") != "1" {
		t.Skip("seeded combo operations MySQL environment is not enabled")
	}
	conn := sqlx.NewMysql(dsn)
	result, err := QueryOptionOperationsOverview(
		context.Background(), conn, 9, 940, 940,
	)
	if err != nil {
		t.Fatalf("query anomalous tenant overview: %v", err)
	}
	if result.ComboStaleCount != 2 ||
		result.ComboManualReviewCount != 1 ||
		result.OldestComboExceptionTime != 800 ||
		result.ComboInvariantIssueCount != 3 ||
		result.ComboIncompleteMatchCount != 1 {
		t.Fatalf("unexpected anomalous tenant combo counts: %+v", result)
	}

	healthy, err := QueryOptionOperationsOverview(
		context.Background(), conn, 10, 940, 940,
	)
	if err != nil {
		t.Fatalf("query healthy tenant overview: %v", err)
	}
	if healthy.ComboStaleCount != 0 ||
		healthy.ComboManualReviewCount != 0 ||
		healthy.OldestComboExceptionTime != 0 ||
		healthy.ComboInvariantIssueCount != 0 ||
		healthy.ComboIncompleteMatchCount != 0 {
		t.Fatalf("unexpected healthy tenant combo counts: %+v", healthy)
	}
}

func TestComboRuntimeBarrierSeededMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_OPERATIONS_TEST_DSN")
	if dsn == "" || os.Getenv("OPTION_COMBO_RUNTIME_EXPECT_SEEDED") != "1" {
		t.Skip("seeded combo runtime MySQL environment is not enabled")
	}
	conn := sqlx.NewMysql(dsn)
	model := &customTOptionOutboxModel{
		defaultTOptionOutboxModel: &defaultTOptionOutboxModel{
			CachedConn: sqlc.NewConnWithCache(conn, nil),
			table:      "`t_option_outbox`",
		},
	}
	tests := []struct {
		name         string
		tenantID     int64
		comboMatchNo string
		want         bool
	}{
		{name: "healthy", tenantID: 10, comboMatchNo: "MATCH-HEALTHY", want: true},
		{name: "complete", tenantID: 9, comboMatchNo: "MATCH-GOOD", want: true},
		{name: "duplicate leg number", tenantID: 9, comboMatchNo: "MATCH-BAD", want: false},
		{name: "missing debit", tenantID: 9, comboMatchNo: "MATCH-BLOCKED", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := model.ComboDebitBarrierReady(
				context.Background(), test.tenantID, test.comboMatchNo,
			)
			if err != nil {
				t.Fatalf("query barrier: %v", err)
			}
			if got != test.want {
				t.Fatalf("ready=%v want=%v", got, test.want)
			}
		})
	}
	blocked, err := model.CountStaleComboDebitBarrierBlocked(
		context.Background(), 9, 940,
	)
	if err != nil {
		t.Fatalf("count anomalous tenant stale barriers: %v", err)
	}
	if blocked != 2 {
		t.Fatalf("anomalous tenant stale barrier events=%d want=2", blocked)
	}
	healthyBlocked, err := model.CountStaleComboDebitBarrierBlocked(
		context.Background(), 10, 940,
	)
	if err != nil {
		t.Fatalf("count healthy tenant stale barriers: %v", err)
	}
	if healthyBlocked != 0 {
		t.Fatalf("healthy tenant stale barrier events=%d want=0", healthyBlocked)
	}
}

func TestQueryOptionOperationsMetricsSeededMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_OPERATIONS_TEST_DSN")
	if dsn == "" || os.Getenv("OPTION_COMBO_RUNTIME_EXPECT_SEEDED") != "1" {
		t.Skip("seeded operations metrics MySQL environment is not enabled")
	}
	counts, amounts, err := QueryOptionOperationsMetrics(
		context.Background(), sqlx.NewMysql(dsn), 940, 940, 1000,
	)
	if err != nil {
		t.Fatalf("query grouped operations metrics: %v", err)
	}
	assertOperationsMetric(t, counts, 9, "asset_pending", 1, 800)
	assertOperationsMetric(t, counts, 9, "outbox_pending", 2, 800)
	assertOperationsMetric(t, counts, 9, "combo_stale", 2, 800)
	assertOperationsMetric(t, counts, 9, "combo_manual_review", 1, 850)
	assertOperationsMetric(t, counts, 9, "combo_invariant_issue", 3, 0)
	assertOperationsMetric(t, counts, 9, "combo_incomplete_match", 1, 0)
	assertOperationsMetric(t, counts, 9, "trading_control_unconfigured", 1, 600)
	assertOperationsMetric(t, counts, 9, "control_not_configured_recent", 1, 950)
	assertOperationsMetric(t, counts, 9, "circuit_breaker_recent", 1, 910)
	assertOperationsMetric(t, counts, 9, "trade_correction_exception", 2, 800)
	assertOperationsMetric(t, counts, 9, "price_band_contract_breach", 1, 950)
	assertOperationsMetric(t, counts, 9, "stp_user_breach", 1, 950)
	assertOperationsMetric(t, counts, 9, "stp_tenant_breach", 1, 950)
	assertOperationsMetric(t, counts, 9, "position_limit_group_breach", 1, 950)
	assertOperationsMetric(t, counts, 9, "paused_contract_active_orders", 1, 800)
	assertOperationsMetric(t, counts, 9, "kill_switch_active_orders", 1, 800)
	assertOperationsMetric(t, counts, 9, "mmp_exception", 2, 800)
	assertOperationsMetric(t, counts, 9, "underlying_market_stale", 1, 900)
	assertOperationsMetric(t, counts, 9, "mark_market_stale", 1, 900)
	assertOperationsMetric(t, counts, 9, "greeks_market_invalid", 1, 0)
	for _, item := range counts {
		if item.TenantID == 10 {
			t.Fatalf("healthy tenant emitted non-zero count metric: %+v", item)
		}
	}
	foundInsurance := false
	for _, item := range amounts {
		if item.TenantID == 9 &&
			item.Category == "insurance_ledger" &&
			item.Coin == "USDT" &&
			item.Amount.Equal(decimal.NewFromInt(7)) {
			foundInsurance = true
		}
	}
	if !foundInsurance {
		t.Fatalf("insurance metric not found in %+v", amounts)
	}
}

func TestQueryOptionGovernanceMetricsSeededMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_OPERATIONS_TEST_DSN")
	if dsn == "" || os.Getenv("OPTION_GOVERNANCE_METRICS_EXPECT_SEEDED") != "1" {
		t.Skip("seeded governance metrics MySQL environment is not enabled")
	}
	counts, err := queryOptionGovernanceMetricsByTenant(
		context.Background(), sqlx.NewMysql(dsn), 100000,
	)
	if err != nil {
		t.Fatalf("query governance metrics: %v", err)
	}
	expected := []struct {
		category string
		count    int64
		oldest   int64
	}{
		{"settlement_price_overdue", 1, 99800},
		{"settlement_price_invalid", 1, 99900},
		{"portfolio_risk_config_missing", 1, 0},
		{"portfolio_risk_version_mismatch", 1, 99800},
		{"corporate_action_due", 1, 99800},
		{"corporate_action_exception", 1, 99700},
		{"contract_series_review_stale", 1, 1000},
		{"contract_series_invariant_issue", 1, 90000},
		{"public_chain_pair_issue", 1, 0},
		{"open_interest_imbalance", 1, 0},
	}
	for _, item := range expected {
		assertOperationsMetric(t, counts, 9, item.category, item.count, item.oldest)
	}
	for _, item := range counts {
		if item.TenantID == 10 {
			t.Fatalf("healthy tenant emitted governance metric: %+v", item)
		}
	}
}

func assertOperationsMetric(
	t *testing.T,
	items []*OptionOperationsMetric,
	tenantID int64,
	category string,
	count, oldest int64,
) {
	t.Helper()
	for _, item := range items {
		if item.TenantID == tenantID && item.Category == category {
			if item.Count != count || item.Oldest != oldest {
				t.Fatalf("%s=%+v want count/oldest=%d/%d", category, item, count, oldest)
			}
			return
		}
	}
	t.Fatalf("metric tenant/category=%d/%s not found in %+v", tenantID, category, items)
}

func assertCoinAmount(t *testing.T, items []*OptionCoinAmount, coin, amount string) {
	t.Helper()
	for _, item := range items {
		if item.Coin == coin {
			if !item.Amount.Equal(decimal.RequireFromString(amount)) {
				t.Fatalf("%s amount=%s want=%s", coin, item.Amount, amount)
			}
			return
		}
	}
	t.Fatalf("coin %s not found in %+v", coin, items)
}
