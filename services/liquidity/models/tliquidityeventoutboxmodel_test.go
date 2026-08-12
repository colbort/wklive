package models

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestLiquidityOutboxLeaseOwnerFencesTransitions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := &customTLiquidityEventOutboxModel{
		CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
		table:      "`t_liquidity_event_outbox`",
	}
	mock.ExpectExec(`(?s)UPDATE .*SET status=2,claimed_by=\?,claimed_at=\?,update_times=\?`).
		WithArgs("worker-a", int64(100), int64(100), int64(9), int64(100), int64(40)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	claimed, err := model.Claim(context.Background(), 9, "worker-a", 100, 40)
	if err != nil || !claimed {
		t.Fatalf("claim=%v err=%v", claimed, err)
	}
	mock.ExpectExec(`(?s)UPDATE .*SET status=3.*claimed_by=''`).
		WithArgs(int64(120), int64(120), int64(9), "worker-b").
		WillReturnResult(sqlmock.NewResult(0, 0))
	updated, err := model.MarkSuccess(context.Background(), 9, "worker-b", 120)
	if err != nil || updated {
		t.Fatalf("wrong owner updated=%v err=%v", updated, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
