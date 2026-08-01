package tasklogic

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestTruncateReconciliationDetailPreservesUTF8(t *testing.T) {
	input := strings.Repeat("差", 1001)
	got := truncateReconciliationDetail(input)
	if !utf8.ValidString(got) {
		t.Fatal("truncated detail is not valid UTF-8")
	}
	if len([]rune(got)) != 1000 {
		t.Fatalf("rune length=%d want=1000", len([]rune(got)))
	}
	if got := truncateReconciliationDetail("  ok  "); got != "ok" {
		t.Fatalf("trimmed detail=%q want ok", got)
	}
}

func TestReconcileFullFundsMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_DAILY_RECONCILIATION_TASK_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_DAILY_RECONCILIATION_TASK_TEST_DSN is not set")
	}
	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	logic := &ProcessDailyReconciliationLogic{ctx: ctx, svcCtx: &svc.ServiceContext{DB: conn}}
	now := time.Date(2026, 8, 1, 0, 5, 0, 0, time.UTC)
	if err := logic.reconcileFullFunds(99, now, true); err != nil {
		t.Fatalf("healthy full-funds run: %v", err)
	}
	assertScope2Run(t, conn, 1, models.OptionReconciliationRunSucceeded, 3, 0)
	if _, err := conn.ExecCtx(ctx, "UPDATE t_option_reconciliation_run_detail SET detail='tampered' WHERE run_id=1"); err == nil {
		t.Fatal("immutable scope-2 detail accepted update")
	}
	if _, err := conn.ExecCtx(ctx, `UPDATE t_user_asset SET total_amount=106,available_amount=106
WHERE tenant_id=99 AND user_id=1001 AND wallet_type=5 AND coin='USDT'`); err != nil {
		t.Fatalf("seed mismatch: %v", err)
	}
	if err := logic.reconcileFullFunds(99, now, true); err != nil {
		t.Fatalf("mismatch full-funds run: %v", err)
	}
	assertScope2Run(t, conn, 2, models.OptionReconciliationRunMismatch, 3, 1)
	var openIssues int64
	if err := conn.QueryRowCtx(ctx, &openIssues, `SELECT COUNT(1) FROM t_option_reconciliation_issue
WHERE tenant_id=99 AND issue_key='DAILY:2026-07-31:1:USDT:2' AND status=1`); err != nil || openIssues != 1 {
		t.Fatalf("open issues=%d err=%v", openIssues, err)
	}
	if _, err := conn.ExecCtx(ctx, `UPDATE t_user_asset SET total_amount=105,available_amount=105
WHERE tenant_id=99 AND user_id=1001 AND wallet_type=5 AND coin='USDT'`); err != nil {
		t.Fatalf("repair mismatch: %v", err)
	}
	if err := logic.reconcileFullFunds(99, now, true); err != nil {
		t.Fatalf("recovered full-funds run: %v", err)
	}
	assertScope2Run(t, conn, 3, models.OptionReconciliationRunSucceeded, 3, 0)
	var resolvedIssues int64
	if err := conn.QueryRowCtx(ctx, &resolvedIssues, `SELECT COUNT(1) FROM t_option_reconciliation_issue
WHERE tenant_id=99 AND issue_key='DAILY:2026-07-31:1:USDT:2' AND status=2 AND resolved_at>0`); err != nil || resolvedIssues != 1 {
		t.Fatalf("resolved issues=%d err=%v", resolvedIssues, err)
	}
}

func assertScope2Run(t *testing.T, conn sqlx.SqlConn, attempt, status, details, mismatches int64) {
	t.Helper()
	var got struct {
		Status, DetailCount, MismatchCount int64
	}
	err := conn.QueryRowCtx(context.Background(), &got, `SELECT r.status,
  (SELECT COUNT(1) FROM t_option_reconciliation_run_detail d WHERE d.run_id=r.id) detail_count,
  r.mismatch_count
FROM t_option_reconciliation_run r
WHERE r.tenant_id=99 AND r.business_date='2026-07-31' AND r.scope=2 AND r.attempt_no=?`, attempt)
	if err != nil || got.Status != status || got.DetailCount != details || got.MismatchCount != mismatches {
		t.Fatalf("attempt=%d run=%+v err=%v", attempt, got, err)
	}
}

func TestReconciliationDayWindowUsesPreviousUTCDate(t *testing.T) {
	now := time.Date(2026, 8, 1, 8, 5, 15, 0, time.FixedZone("UTC+8", 8*60*60))
	start, end := reconciliationDayWindow(now)
	if got := time.UnixMilli(start).UTC().Format(time.RFC3339); got != "2026-07-31T00:00:00Z" {
		t.Fatalf("start=%s", got)
	}
	if got := time.UnixMilli(end).UTC().Format(time.RFC3339); got != "2026-08-01T00:00:00Z" {
		t.Fatalf("end=%s", got)
	}
}

func TestConservationDetailStatusFailsClosed(t *testing.T) {
	if got := conservationDetailStatus(decimal.Zero, true); got != models.OptionReconciliationDetailIncomplete {
		t.Fatalf("incomplete status=%d", got)
	}
	if got := conservationDetailStatus(decimal.NewFromInt(1), false); got != models.OptionReconciliationDetailMismatch {
		t.Fatalf("mismatch status=%d", got)
	}
	if got := conservationDetailStatus(decimal.Zero, false); got != models.OptionReconciliationDetailMatched {
		t.Fatalf("matched status=%d", got)
	}
}

func TestReconciliationBusinessDateUsesPreviousUTCDay(t *testing.T) {
	now := time.Date(2026, 8, 1, 8, 5, 15, 0, time.FixedZone("UTC+8", 8*60*60))
	if got := reconciliationBusinessDate(now); got != "2026-07-31" {
		t.Fatalf("business date=%s want 2026-07-31", got)
	}
}
