package models

import (
	"context"
	"fmt"
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

func TestQueryOptionControlMetricsPriceBandRatioMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_CONTROL_RATIO_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_CONTROL_RATIO_TEST_DSN is not set")
	}
	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	const now int64 = 1000
	seedEvaluations := func(tenantID, contractID, requestCount, rejectedCount int64) {
		t.Helper()
		for index := int64(0); index < requestCount; index++ {
			if _, err := conn.ExecCtx(ctx, `INSERT INTO t_option_trading_control_event
(tenant_id,user_id,contract_id,event_type,reason,detail,operator_id,create_times)
VALUES(?,1,?,'ORDER_CONTROL_EVALUATED','TRADING_CONTROL','accepted',1,950)`,
				tenantID, contractID); err != nil {
				t.Fatalf("seed evaluation tenant=%d: %v", tenantID, err)
			}
		}
		for index := int64(0); index < rejectedCount; index++ {
			if _, err := conn.ExecCtx(ctx, `INSERT INTO t_option_trading_control_event
(tenant_id,user_id,contract_id,event_type,reason,detail,operator_id,create_times)
VALUES(?,1,?,'ORDER_REJECTED','ORDER_PRICE_BAND','outside band',1,951)`,
				tenantID, contractID); err != nil {
				t.Fatalf("seed rejection tenant=%d: %v", tenantID, err)
			}
		}
	}
	seedEvaluations(91, 9101, 10, 1)  // exactly 10%, no breach
	seedEvaluations(92, 9201, 10, 2)  // greater than 10%
	seedEvaluations(93, 9301, 21, 21) // absolute threshold and ratio
	seedEvaluations(94, 9401, 20, 2)  // exactly 10%, no breach
	seedEvaluations(95, 9501, 0, 21)  // legacy window: absolute threshold still works

	metrics, err := queryOptionControlMetricsByTenant(ctx, conn, now)
	if err != nil {
		t.Fatalf("query price-band ratio metrics: %v", err)
	}
	breaches := make(map[int64]*OptionOperationsMetric)
	for _, metric := range metrics {
		if metric.Category == "price_band_contract_breach" {
			breaches[metric.TenantID] = metric
		}
	}
	if len(breaches) != 3 || breaches[92] == nil || breaches[93] == nil || breaches[95] == nil {
		t.Fatalf("price-band breaches=%+v want tenants 92, 93 and 95", breaches)
	}
	for _, tenantID := range []int64{92, 93, 95} {
		if breaches[tenantID].Count != 1 || breaches[tenantID].Oldest != 951 {
			t.Fatalf("tenant=%d breach=%+v", tenantID, breaches[tenantID])
		}
	}
}

func TestQueryOptionMarketMetricsGreeksThresholdMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_GREEKS_FRESHNESS_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_GREEKS_FRESHNESS_TEST_DSN is not set")
	}
	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	const now int64 = 1000
	tenantIDs := []int64{101, 102, 103, 104, 105}
	cleanup := func() {
		for _, tenantID := range tenantIDs {
			_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_market WHERE tenant_id=?", tenantID)
			_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_contract WHERE tenant_id=?", tenantID)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	seed := func(tenantID, threshold, snapshot int64, withMarket bool) {
		t.Helper()
		result, err := conn.ExecCtx(ctx, `INSERT INTO t_option_contract
(tenant_id,contract_code,status,greeks_max_age_seconds)
VALUES(?,?,2,?)`, tenantID, fmt.Sprintf("GREEKS-%d", tenantID), threshold)
		if err != nil {
			t.Fatalf("seed contract tenant=%d: %v", tenantID, err)
		}
		if !withMarket {
			return
		}
		contractID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("contract id tenant=%d: %v", tenantID, err)
		}
		if _, err = conn.ExecCtx(ctx, `INSERT INTO t_option_market
(tenant_id,contract_id,greeks_snapshot_time)
VALUES(?,?,?)`, tenantID, contractID, snapshot); err != nil {
			t.Fatalf("seed market tenant=%d: %v", tenantID, err)
		}
	}
	seed(101, 10, 990, true)  // exact boundary: fresh
	seed(102, 10, 989, true)  // older than approved threshold
	seed(103, 0, 1000, true)  // threshold not configured
	seed(104, 10, 1001, true) // future timestamp
	seed(105, 10, 0, false)   // market row missing

	metrics, err := queryOptionMarketMetricsByTenant(ctx, conn, now)
	if err != nil {
		t.Fatalf("query market metrics: %v", err)
	}
	greeks := make(map[int64]*OptionOperationsMetric)
	for _, metric := range metrics {
		if metric.Category == "greeks_market_invalid" {
			greeks[metric.TenantID] = metric
		}
	}
	if greeks[101] != nil {
		t.Fatalf("boundary-fresh tenant emitted alert: %+v", greeks[101])
	}
	assertOperationsMetric(t, metrics, 102, "greeks_market_invalid", 1, 989)
	assertOperationsMetric(t, metrics, 103, "greeks_market_invalid", 1, 0)
	assertOperationsMetric(t, metrics, 104, "greeks_market_invalid", 1, 0)
	assertOperationsMetric(t, metrics, 105, "greeks_market_invalid", 1, 0)
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

	allCounts, _, err := QueryOptionOperationsMetrics(
		context.Background(), sqlx.NewMysql(dsn), 99940, 99940, 100000,
	)
	if err != nil {
		t.Fatalf("query all operations metrics: %v", err)
	}
	assertOperationsMetric(t, allCounts, 9, "daily_conservation_issue", 1, 99950)
}

func TestQueryOptionRemainingAlertMetricsSeededMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_OPERATIONS_TEST_DSN")
	if dsn == "" || os.Getenv("OPTION_REMAINING_ALERTS_EXPECT_SEEDED") != "1" {
		t.Skip("seeded remaining-alert metrics MySQL environment is not enabled")
	}
	counts, _, err := QueryOptionOperationsMetrics(
		context.Background(), sqlx.NewMysql(dsn), 99940, 99940, 100000,
	)
	if err != nil {
		t.Fatalf("query remaining alert metrics: %v", err)
	}
	assertOperationsMetric(t, counts, 9, "exercise_near_expiry", 1, 99900)
	assertOperationsMetric(t, counts, 9, "kill_switch_release_failure", 1, 99920)
	assertOperationsMetric(t, counts, 9, "physical_delivery_exception", 2, 99800)
	assertOperationsMetric(t, counts, 9, "physical_delivery_overdue", 1, 99900)
	for _, item := range counts {
		if item.TenantID == 10 &&
			(item.Category == "exercise_near_expiry" ||
				item.Category == "kill_switch_release_failure" ||
				item.Category == "physical_delivery_exception" ||
				item.Category == "physical_delivery_overdue") {
			t.Fatalf("healthy tenant emitted remaining-alert metric: %+v", item)
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
