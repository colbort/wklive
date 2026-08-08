package models

import (
	"context"
	"testing"

	"wklive/proto/option"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestOptionOutboxLeaseOwnerFencesFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := &customTOptionOutboxModel{
		defaultTOptionOutboxModel: &defaultTOptionOutboxModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_outbox`",
		},
	}
	mock.ExpectExec(`(?s)UPDATE .*SET status=\?,claimed_by=\?,claimed_at=\?,update_times=\?`).
		WithArgs(
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PROCESSING), "worker-a",
			int64(100), int64(100), int64(8),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED), int64(100),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PROCESSING), int64(40),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	claimed, err := model.Claim(context.Background(), 8, "worker-a", 100, 40)
	if err != nil || !claimed {
		t.Fatalf("claim=%v err=%v", claimed, err)
	}
	mock.ExpectExec(`(?s)UPDATE .*SET status=\?,retry_count=\?.*claimed_by=''`).
		WithArgs(
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED), int64(1), int64(130),
			"failed", int64(110), int64(8),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PROCESSING), "worker-b",
		).
		WillReturnResult(sqlmock.NewResult(0, 0))
	updated, err := model.MarkFailed(
		context.Background(), 8, "worker-b", 1,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED), 130, "failed", 110,
	)
	if err != nil || updated {
		t.Fatalf("wrong owner updated=%v err=%v", updated, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
