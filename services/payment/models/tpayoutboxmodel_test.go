package models

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"wklive/proto/payment"
)

func TestClaimPendingUsesCompareAndSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := newOutboxSQLMockModel(db)

	const (
		now         = int64(10_000)
		staleBefore = int64(1_000)
		rowID       = int64(7)
	)
	for attempt, affected := range []int64{1, 0} {
		mock.ExpectQuery(`(?s)select .* from .* where .*status.*next_retry_at.*claimed_at.*order by`).
			WithArgs(now, staleBefore, int64(10)).
			WillReturnRows(outboxRows(rowID))
		mock.ExpectExec(`(?s)update .*set status = \?, claimed_by = \?, claimed_at = \?.*where id = \?.*status in`).
			WithArgs(
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PROCESSING),
				"worker-a",
				now,
				now,
				rowID,
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PENDING),
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_FAILED),
				now,
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PROCESSING),
				staleBefore,
			).
			WillReturnResult(sqlmock.NewResult(0, affected))

		claimed, err := model.ClaimPending(context.Background(), "worker-a", now, staleBefore, 10)
		if err != nil {
			t.Fatalf("attempt %d claim failed: %v", attempt+1, err)
		}
		if got, want := len(claimed), int(affected); got != want {
			t.Fatalf("attempt %d claimed=%d want=%d", attempt+1, got, want)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOutboxCompletionRequiresClaimOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := newOutboxSQLMockModel(db)
	row := &TPayOutbox{Id: 7, EventNo: "evt-7"}

	pattern := regexp.QuoteMeta("update `t_pay_outbox` set status = ?, claimed_by = '', claimed_at = 0, next_retry_at = 0, last_error_msg = '', update_times = ? where id = ? and status = ? and claimed_by = ?")
	for attempt, affected := range []int64{1, 0} {
		mock.ExpectExec(pattern).
			WithArgs(
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_SUCCESS),
				int64(20_000),
				row.Id,
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PROCESSING),
				"worker-a",
			).
			WillReturnResult(sqlmock.NewResult(0, affected))
		updated, err := model.MarkSuccess(context.Background(), row, "worker-a", 20_000)
		if err != nil {
			t.Fatalf("attempt %d mark success failed: %v", attempt+1, err)
		}
		if updated != (affected == 1) {
			t.Fatalf("attempt %d updated=%t affected=%d", attempt+1, updated, affected)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkFailedClearsClaimAndTruncatesError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := newOutboxSQLMockModel(db)
	row := &TPayOutbox{Id: 8, EventNo: "evt-8"}
	message := strings.Repeat("错", 1_001)

	mock.ExpectExec(`(?s)update .*set status = \?, retry_count = \?, next_retry_at = \?, last_error_msg = \?, claimed_by = ''.*where id = \?.*claimed_by = \?`).
		WithArgs(
			int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_FAILED),
			int64(3),
			int64(30_000),
			truncatePayOutboxError(message),
			int64(20_000),
			row.Id,
			int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PROCESSING),
			"worker-b",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	updated, err := model.MarkFailed(context.Background(), row, "worker-b", 3, 30_000, 20_000, message)
	if err != nil || !updated {
		t.Fatalf("mark failed=(%t,%v), want true,nil", updated, err)
	}
	if got := len([]rune(truncatePayOutboxError(message))); got != 1_000 {
		t.Fatalf("truncated runes=%d want=1000", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func newOutboxSQLMockModel(db *sql.DB) *customTPayOutboxModel {
	return &customTPayOutboxModel{
		defaultTPayOutboxModel: &defaultTPayOutboxModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), noopCache{}),
			table:      "`t_pay_outbox`",
		},
	}
}

var errNoopCacheMiss = errors.New("cache miss")

type noopCache struct{}

func (noopCache) Del(...string) error                            { return nil }
func (noopCache) DelCtx(context.Context, ...string) error        { return nil }
func (noopCache) Get(string, any) error                          { return errNoopCacheMiss }
func (noopCache) GetCtx(context.Context, string, any) error      { return errNoopCacheMiss }
func (noopCache) IsNotFound(err error) bool                      { return errors.Is(err, errNoopCacheMiss) }
func (noopCache) Set(string, any) error                          { return nil }
func (noopCache) SetCtx(context.Context, string, any) error      { return nil }
func (noopCache) SetWithExpire(string, any, time.Duration) error { return nil }
func (noopCache) SetWithExpireCtx(context.Context, string, any, time.Duration) error {
	return nil
}
func (noopCache) Take(value any, _ string, query func(any) error) error {
	return query(value)
}
func (noopCache) TakeCtx(_ context.Context, value any, _ string, query func(any) error) error {
	return query(value)
}
func (noopCache) TakeWithExpire(value any, _ string, query func(any, time.Duration) error) error {
	return query(value, 0)
}
func (noopCache) TakeWithExpireCtx(_ context.Context, value any, _ string, query func(any, time.Duration) error) error {
	return query(value, 0)
}

func outboxRows(id int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "event_no", "event_type", "aggregate_type", "aggregate_id", "aggregate_no",
		"payload", "status", "claimed_by", "claimed_at", "retry_count", "next_retry_at",
		"last_error_msg", "create_times", "update_times",
	}).AddRow(
		id, "evt-7", "PAYMENT_RECHARGE_CREDIT", "RECHARGE_ORDER", 99, "PAY-99",
		`{"tenantId":1}`, 1, "", 0, 0, 0, "", 1, 1,
	)
}
