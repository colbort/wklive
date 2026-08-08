package models

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestSnapshotOutboxClaimAndCompletionAreOwnerFenced(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := &customTItickSnapshotOutboxModel{
		defaultTItickSnapshotOutboxModel: &defaultTItickSnapshotOutboxModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_itick_snapshot_outbox`",
		},
	}
	mock.ExpectExec(`(?s)UPDATE t_itick_snapshot_outbox SET status=2,claimed_by=\?,claimed_at=\?,update_times=\?.*`).
		WithArgs("worker-a", int64(100), int64(100), int64(7), int64(100), int64(40)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	claimed, err := model.Claim(context.Background(), 7, "worker-a", 100, 40)
	if err != nil || !claimed {
		t.Fatalf("claim=%v err=%v", claimed, err)
	}
	mock.ExpectExec(`(?s)UPDATE t_itick_snapshot_outbox SET status=3.*claimed_by=''`).
		WithArgs(int64(110), int64(7), "worker-b").
		WillReturnResult(sqlmock.NewResult(0, 0))
	err = model.MarkSuccess(context.Background(), 7, "worker-b", 110)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("wrong owner completion err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
