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

func TestStakeOperationVersionFencesStaleWorker(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := &customTStakeOperationModel{
		defaultTStakeOperationModel: &defaultTStakeOperationModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_stake_operation`",
		},
	}
	mock.ExpectExec(`(?s)UPDATE t_stake_operation.*version=version\+1.*WHERE id=\? AND status=2 AND version=\?`).
		WithArgs(int64(2), int64(1), int64(0), int64(100), int64(7), int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	_, err = model.CheckpointSteps(context.Background(), 7, 5, 2, 1, 0, 100)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("stale checkpoint err=%v", err)
	}
	mock.ExpectExec(`(?s)UPDATE t_stake_operation.*WHERE id=\? AND status=2 AND version=\?`).
		WithArgs(int64(4), int64(1), int64(130), "failed", int64(110), int64(7), int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err = model.MarkRetryable(context.Background(), 7, 5, 1, 130, 4, 110, "failed")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("stale failure transition err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
