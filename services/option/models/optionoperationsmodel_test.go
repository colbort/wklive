package models

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

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
