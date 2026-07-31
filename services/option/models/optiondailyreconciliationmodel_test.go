package models

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestListOptionReconciliationTenantIDsHonorsExplicitTenant(t *testing.T) {
	got, err := ListOptionReconciliationTenantIDs(context.Background(), nil, 900101)
	if err != nil {
		t.Fatalf("list explicit tenant: %v", err)
	}
	if len(got) != 1 || got[0] != 900101 {
		t.Fatalf("tenant IDs=%v want [900101]", got)
	}
}

func TestQueryOptionAccountMirrorSummariesMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_DAILY_RECONCILIATION_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_DAILY_RECONCILIATION_TEST_DSN is not set")
	}
	conn := sqlx.NewMysql(dsn)
	if _, err := conn.ExecCtx(context.Background(), `INSERT INTO t_option_account
(tenant_id,user_id,account_id,margin_coin,balance,available_balance,frozen_balance,status)
VALUES(11,1,0,'ETH',0.000000000000000001,0.000000000000000001,0,1)
ON DUPLICATE KEY UPDATE balance=VALUES(balance),available_balance=VALUES(available_balance),
frozen_balance=VALUES(frozen_balance)`); err != nil {
		t.Fatalf("seed high-precision Option mirror: %v", err)
	}
	if _, err := conn.ExecCtx(context.Background(), `INSERT INTO t_user_asset
(tenant_id,user_id,wallet_type,coin,total_amount,available_amount,frozen_amount,locked_amount)
VALUES(11,1,5,'ETH',0.000000000000000001,0.000000000000000001,0,0)
ON DUPLICATE KEY UPDATE total_amount=VALUES(total_amount),available_amount=VALUES(available_amount),
frozen_amount=VALUES(frozen_amount),locked_amount=VALUES(locked_amount)`); err != nil {
		t.Fatalf("seed high-precision Asset wallet: %v", err)
	}
	rows, err := QueryOptionAccountMirrorSummaries(context.Background(), conn, 9)
	if err != nil {
		t.Fatalf("query tenant 9 mirror summaries: %v", err)
	}
	expected := map[string]struct {
		accounts   int64
		mismatches int64
	}{
		"BTC":  {accounts: 2, mismatches: 2},
		"ETH":  {accounts: 1, mismatches: 1},
		"USDT": {accounts: 2, mismatches: 1},
	}
	if len(rows) != len(expected) {
		t.Fatalf("tenant 9 summaries=%+v", rows)
	}
	for _, row := range rows {
		want, ok := expected[row.Coin]
		if !ok || row.AccountCount != want.accounts || row.MismatchCount != want.mismatches {
			t.Fatalf("tenant 9 summary=%+v expected=%+v", row, want)
		}
	}
	healthy, err := QueryOptionAccountMirrorSummaries(context.Background(), conn, 10)
	if err != nil {
		t.Fatalf("query tenant 10 mirror summaries: %v", err)
	}
	if len(healthy) != 1 || healthy[0].Coin != "USDT" || healthy[0].MismatchCount != 0 {
		t.Fatalf("healthy summaries=%+v", healthy)
	}
	highPrecision, err := QueryOptionAccountMirrorSummaries(context.Background(), conn, 11)
	if err != nil {
		t.Fatalf("query high-precision mirror summaries: %v", err)
	}
	unit := decimal.RequireFromString("0.000000000000000001")
	if len(highPrecision) != 1 || highPrecision[0].MismatchCount != 0 ||
		!highPrecision[0].OptionTotal.Equal(unit) || !highPrecision[0].AssetTotal.Equal(unit) {
		t.Fatalf("high-precision summaries=%+v", highPrecision)
	}

	for _, run := range []*OptionReconciliationRun{
		{
			TenantID: 9, BusinessDate: "2026-07-31",
			Scope:  OptionReconciliationScopeAccountMirror,
			Status: OptionReconciliationRunMismatch, SnapshotTime: 100000,
			SnapshotRef: "mysql-single-statement:100000", CoinCount: 3,
			AccountCount: 5, MismatchCount: 4, Detail: "seeded mismatch", CompletedAt: 100000,
		},
		{
			TenantID: 10, BusinessDate: "2026-07-31",
			Scope:  OptionReconciliationScopeAccountMirror,
			Status: OptionReconciliationRunSucceeded, SnapshotTime: 100000,
			SnapshotRef: "mysql-single-statement:100000", CoinCount: 1,
			AccountCount: 1, Detail: "seeded success", CompletedAt: 100000,
		},
	} {
		attempt, attemptErr := NextOptionReconciliationAttempt(
			context.Background(), conn, run.TenantID, run.BusinessDate, run.Scope,
		)
		if attemptErr != nil {
			t.Fatalf("next MySQL attempt: %v", attemptErr)
		}
		run.AttemptNo = attempt
		if insertErr := InsertOptionReconciliationRun(context.Background(), conn, run); insertErr != nil {
			t.Fatalf("insert MySQL run: %v", insertErr)
		}
	}
	metrics, err := queryOptionDailyReconciliationMetricsByTenant(
		context.Background(), conn, 100000,
	)
	if err != nil {
		t.Fatalf("query daily reconciliation metrics: %v", err)
	}
	assertDailyReconciliationMetric(t, metrics, 9, "daily_mirror_reconciliation_heartbeat", 0, 0)
	assertDailyReconciliationMetric(t, metrics, 9, "daily_mirror_reconciliation_mismatch", 1, 100000)
	assertDailyReconciliationMetric(t, metrics, 9, "daily_mirror_reconciliation_missing", 1, 0)
	assertDailyReconciliationMetric(t, metrics, 9, "daily_conservation_heartbeat", 0, 0)
	assertDailyReconciliationMetric(t, metrics, 9, "daily_conservation_heartbeat_missing", 1, 0)
	assertDailyReconciliationMetric(t, metrics, 10, "daily_mirror_reconciliation_heartbeat", 0, 100000)
	assertDailyReconciliationMetric(t, metrics, 10, "daily_conservation_heartbeat", 0, 0)
	assertDailyReconciliationMetric(t, metrics, 10, "daily_conservation_heartbeat_missing", 1, 0)

	if _, updateErr := conn.ExecCtx(context.Background(),
		"UPDATE t_option_reconciliation_run SET detail='tampered' WHERE tenant_id=10"); updateErr == nil {
		t.Fatal("immutable reconciliation run accepted UPDATE")
	}
	if _, deleteErr := conn.ExecCtx(context.Background(),
		"DELETE FROM t_option_reconciliation_run WHERE tenant_id=10"); deleteErr == nil {
		t.Fatal("immutable reconciliation run accepted DELETE")
	}
}

func TestQueryOptionUserWalletConservationSummariesMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_DAILY_RECONCILIATION_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_DAILY_RECONCILIATION_TEST_DSN is not set")
	}
	conn := sqlx.NewMysql(dsn)
	ctx := context.Background()
	rows, err := QueryOptionUserWalletConservationSummaries(ctx, conn, 9, 1000, 2000, 2500)
	if err != nil {
		t.Fatalf("query healthy user-wallet conservation: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("healthy rows=%+v", rows)
	}
	row := rows[0]
	assertConservationDecimal(t, "opening", row.OpeningAmount, "150")
	assertConservationDecimal(t, "external", row.ExternalNet, "10")
	assertConservationDecimal(t, "option", row.OptionNet, "-3")
	assertConservationDecimal(t, "manual", row.ManualNet, "2")
	assertConservationDecimal(t, "expected", row.ExpectedClosing, "159")
	assertConservationDecimal(t, "actual", row.ActualClosing, "159")
	assertConservationDecimal(t, "difference", row.DifferenceAmount, "0")
	if row.Coin != "USDT" || row.WalletCount != 2 || row.MismatchWalletCount != 0 ||
		row.IntegrityErrorCount != 0 || row.UnclassifiedFlowCount != 0 || row.FlowCount != 3 || row.MaxFlowID != 6 {
		t.Fatalf("healthy summary=%+v", row)
	}

	unknown, err := QueryOptionUserWalletConservationSummaries(ctx, conn, 10, 1000, 2000, 2500)
	if err != nil || len(unknown) != 1 || unknown[0].UnclassifiedFlowCount != 1 ||
		unknown[0].MismatchWalletCount != 1 || !unknown[0].DifferenceAmount.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("unknown classification rows=%+v err=%v", unknown, err)
	}
	broken, err := QueryOptionUserWalletConservationSummaries(ctx, conn, 11, 1000, 2000, 2500)
	if err != nil || len(broken) != 1 || broken[0].IntegrityErrorCount == 0 ||
		broken[0].MismatchWalletCount != 1 || !broken[0].DifferenceAmount.IsZero() {
		t.Fatalf("broken chain rows=%+v err=%v", broken, err)
	}
	highPrecision, err := QueryOptionUserWalletConservationSummaries(ctx, conn, 12, 1000, 2000, 2500)
	unit := decimal.RequireFromString("0.000000000000000001")
	if err != nil || len(highPrecision) != 1 || highPrecision[0].MismatchWalletCount != 0 ||
		!highPrecision[0].ExternalNet.Equal(unit) || !highPrecision[0].ActualClosing.Equal(unit) {
		t.Fatalf("high precision rows=%+v err=%v", highPrecision, err)
	}
}

func TestQueryOptionPlatformAccountConservationSummariesMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_DAILY_RECONCILIATION_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_DAILY_RECONCILIATION_TEST_DSN is not set")
	}
	conn := sqlx.NewMysql(dsn)
	ctx := context.Background()
	rows, err := QueryOptionPlatformAccountConservationSummaries(ctx, conn, 9, 1000, 2000, 2500)
	if err != nil || len(rows) != 2 {
		t.Fatalf("healthy platform rows=%+v err=%v", rows, err)
	}
	byType := make(map[string]*OptionPlatformAccountConservationSummary, len(rows))
	for _, row := range rows {
		byType[row.AccountType] = row
	}
	fee := byType["FEE_REVENUE"]
	if fee == nil || fee.MismatchAccountCount != 0 || fee.IntegrityErrorCount != 0 || fee.FlowCount != 2 {
		t.Fatalf("fee summary=%+v", fee)
	}
	assertConservationDecimal(t, "platform opening", fee.OpeningAmount, "100")
	assertConservationDecimal(t, "platform option", fee.OptionNet, "5")
	assertConservationDecimal(t, "platform manual", fee.ManualNet, "1")
	assertConservationDecimal(t, "platform expected", fee.ExpectedClosing, "106")
	assertConservationDecimal(t, "platform actual", fee.ActualClosing, "106")
	insurance := byType["INSURANCE_FUND"]
	if insurance == nil || insurance.MismatchAccountCount != 0 || insurance.FlowCount != 0 ||
		!insurance.ExpectedClosing.Equal(decimal.NewFromInt(50)) {
		t.Fatalf("insurance summary=%+v", insurance)
	}
	broken, err := QueryOptionPlatformAccountConservationSummaries(ctx, conn, 10, 1000, 2000, 2500)
	if err != nil || len(broken) != 1 || broken[0].IntegrityErrorCount == 0 ||
		broken[0].MismatchAccountCount != 1 || !broken[0].DifferenceAmount.IsZero() {
		t.Fatalf("broken platform rows=%+v err=%v", broken, err)
	}
	frozen, err := QueryOptionPlatformAccountConservationSummaries(ctx, conn, 11, 1000, 2000, 2500)
	if err != nil || len(frozen) != 1 || frozen[0].IntegrityErrorCount != 1 || frozen[0].MismatchAccountCount != 1 {
		t.Fatalf("frozen platform rows=%+v err=%v", frozen, err)
	}
	highPrecision, err := QueryOptionPlatformAccountConservationSummaries(ctx, conn, 12, 1000, 2000, 2500)
	unit := decimal.RequireFromString("0.000000000000000001")
	if err != nil || len(highPrecision) != 1 || highPrecision[0].MismatchAccountCount != 0 ||
		!highPrecision[0].OptionNet.Equal(unit) || !highPrecision[0].ActualClosing.Equal(unit) {
		t.Fatalf("high precision platform rows=%+v err=%v", highPrecision, err)
	}
}

func assertConservationDecimal(t *testing.T, name string, got decimal.Decimal, want string) {
	t.Helper()
	if !got.Equal(decimal.RequireFromString(want)) {
		t.Fatalf("%s=%s want %s", name, got, want)
	}
}

func assertDailyReconciliationMetric(
	t *testing.T, items []*OptionOperationsMetric, tenantID int64, category string, count, oldest int64,
) {
	t.Helper()
	for _, item := range items {
		if item.TenantID == tenantID && item.Category == category {
			if item.Count != count || item.Oldest != oldest {
				t.Fatalf("metric=%+v want count=%d oldest=%d", item, count, oldest)
			}
			return
		}
	}
	t.Fatalf("missing metric tenant=%d category=%s in %+v", tenantID, category, items)
}

func TestOptionDailyReconciliationModelQueriesAndAppendsRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("new sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	conn := sqlx.NewSqlConnFromDB(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(MAX(attempt_no),0)+1")).
		WithArgs(int64(9), "2026-07-31", OptionReconciliationScopeAccountMirror).
		WillReturnRows(sqlmock.NewRows([]string{"attempt_no"}).AddRow(3))
	attempt, err := NextOptionReconciliationAttempt(
		ctx, conn, 9, "2026-07-31", OptionReconciliationScopeAccountMirror,
	)
	if err != nil || attempt != 3 {
		t.Fatalf("next attempt=%d err=%v", attempt, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1)\nFROM t_option_reconciliation_run")).
		WithArgs(int64(9), "2026-07-31", OptionReconciliationScopeAccountMirror).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	succeeded, err := HasSuccessfulOptionReconciliationRun(
		ctx, conn, 9, "2026-07-31", OptionReconciliationScopeAccountMirror,
	)
	if err != nil || !succeeded {
		t.Fatalf("has successful run=%v err=%v", succeeded, err)
	}

	mock.ExpectQuery("WITH wallet_keys AS").
		WithArgs(int64(9), int64(9), int64(9), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{
			"coin", "account_count", "mismatch_count",
			"option_total", "option_available", "option_frozen",
			"asset_total", "asset_available", "asset_frozen", "asset_locked",
		}).AddRow("USDT", 2, 1, "10", "7", "3", "11", "8", "3", "0"))
	summaries, err := QueryOptionAccountMirrorSummaries(ctx, conn, 9)
	if err != nil {
		t.Fatalf("query summaries: %v", err)
	}
	if len(summaries) != 1 || summaries[0].Coin != "USDT" ||
		summaries[0].AccountCount != 2 || summaries[0].MismatchCount != 1 ||
		!summaries[0].OptionTotal.Equal(decimal.NewFromInt(10)) ||
		!summaries[0].AssetTotal.Equal(decimal.NewFromInt(11)) {
		t.Fatalf("unexpected summaries: %+v", summaries)
	}
	mock.ExpectQuery("WITH cutoff AS").WithArgs(
		int64(9), int64(2500), int64(9), int64(1000), int64(9), int64(1000), int64(2500),
		int64(1000), int64(2000), int64(1000), int64(2000), int64(1000), int64(2000),
		int64(1000), int64(2000), int64(1000), int64(2000), int64(2000),
		int64(1000), int64(2000), int64(2000), int64(9), int64(9),
	).WillReturnRows(sqlmock.NewRows([]string{
		"coin", "wallet_count", "mismatch_wallet_count", "integrity_error_count",
		"unclassified_flow_count", "flow_count", "max_flow_id", "opening_amount",
		"external_net", "option_net", "manual_net", "expected_closing", "actual_closing",
		"difference_amount",
	}).AddRow("USDT", 2, 0, 0, 0, 3, 6, "150", "10", "-3", "2", "159", "159", "0"))
	conservation, err := QueryOptionUserWalletConservationSummaries(ctx, conn, 9, 1000, 2000, 2500)
	if err != nil || len(conservation) != 1 || conservation[0].WalletCount != 2 ||
		!conservation[0].DifferenceAmount.IsZero() {
		t.Fatalf("conservation=%+v err=%v", conservation, err)
	}
	mock.ExpectQuery("WITH cutoff AS").WithArgs(
		int64(9), int64(2500), int64(9), int64(1000), int64(9), int64(1000), int64(2500),
		int64(1000), int64(2000), int64(1000), int64(2000), int64(1000), int64(2000),
		int64(2000), int64(1000), int64(2000), int64(2000), int64(9), int64(9),
	).WillReturnRows(sqlmock.NewRows([]string{
		"account_type", "coin", "account_count", "mismatch_account_count",
		"integrity_error_count", "flow_count", "max_platform_flow_id", "opening_amount",
		"option_net", "manual_net", "expected_closing", "actual_closing", "difference_amount",
	}).AddRow("FEE_REVENUE", "USDT", 1, 0, 0, 2, 4, "100", "5", "1", "106", "106", "0"))
	platform, err := QueryOptionPlatformAccountConservationSummaries(ctx, conn, 9, 1000, 2000, 2500)
	if err != nil || len(platform) != 1 || platform[0].AccountType != "FEE_REVENUE" ||
		!platform[0].DifferenceAmount.IsZero() {
		t.Fatalf("platform=%+v err=%v", platform, err)
	}

	run := &OptionReconciliationRun{
		TenantID: 9, BusinessDate: "2026-07-31",
		Scope: OptionReconciliationScopeAccountMirror, AttemptNo: 3,
		Status: OptionReconciliationRunMismatch, SnapshotTime: 1000,
		SnapshotRef: "mysql-single-statement:1000", CoinCount: 1,
		AccountCount: 2, MismatchCount: 1, Detail: "mismatch", CompletedAt: 1000,
	}
	mock.ExpectExec("INSERT INTO t_option_reconciliation_run").
		WithArgs(
			run.TenantID, run.BusinessDate, run.Scope, run.AttemptNo, run.Status,
			run.SnapshotTime, run.SnapshotRef, run.CoinCount, run.AccountCount,
			run.MismatchCount, run.Detail, run.CompletedAt, run.CompletedAt, run.CompletedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	runID, err := InsertOptionReconciliationRunWithID(ctx, conn, run)
	if err != nil || runID != 1 {
		t.Fatalf("insert run id=%d err=%v", runID, err)
	}
	detail := &OptionReconciliationRunDetail{
		RunID: 1, TenantID: 9, BusinessDate: "2026-07-31",
		Scope:         OptionReconciliationScopeFullFunds,
		DimensionType: OptionReconciliationDimensionUserWallet,
		DimensionKey:  "USDT", OpeningAmount: decimal.NewFromInt(100),
		ExternalNet: decimal.NewFromInt(5), OptionNet: decimal.NewFromInt(-3),
		ManualNet: decimal.NewFromInt(2), ExpectedClosing: decimal.NewFromInt(104),
		ActualClosing: decimal.NewFromInt(104), DifferenceAmount: decimal.Zero,
		FlowCount: 7, Status: OptionReconciliationDetailMatched,
		EvidenceRef: "asset_flow<=20", Detail: "complete", CreateTimes: 1000,
	}
	mock.ExpectExec("INSERT INTO t_option_reconciliation_run_detail").
		WithArgs(
			detail.RunID, detail.TenantID, detail.BusinessDate, detail.Scope,
			detail.DimensionType, detail.DimensionKey, detail.OpeningAmount,
			detail.ExternalNet, detail.OptionNet, detail.ManualNet,
			detail.ExpectedClosing, detail.ActualClosing, detail.DifferenceAmount,
			detail.FlowCount, detail.MismatchCount, detail.Status,
			detail.EvidenceRef, detail.Detail, detail.CreateTimes,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := InsertOptionReconciliationRunDetail(ctx, conn, detail); err != nil {
		t.Fatalf("insert run detail: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
