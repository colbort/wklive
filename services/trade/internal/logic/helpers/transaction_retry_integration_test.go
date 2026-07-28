package helpers

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// TestDeadlockRetryAgainstMySQL is intentionally opt-in because it creates two
// concurrent real MySQL transactions. Run it only against an isolated database:
//
//	TRADE_MYSQL_TEST_DSN='user:pass@tcp(127.0.0.1:3306)/wklive?parseTime=true' \
//	go test ./internal/logic/helpers -run TestDeadlockRetryAgainstMySQL -count=1
func TestDeadlockRetryAgainstMySQL(t *testing.T) {
	dsn := os.Getenv("TRADE_MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("TRADE_MYSQL_TEST_DSN is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := sqlx.NewMysql(dsn)
	const table = "t_trade_deadlock_acceptance"
	if _, err := db.ExecCtx(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dropCancel()
		_, _ = db.ExecCtx(dropCtx, "DROP TABLE IF EXISTS "+table)
	})
	if _, err := db.ExecCtx(ctx, "CREATE TABLE "+table+" (id BIGINT PRIMARY KEY, value BIGINT NOT NULL) ENGINE=InnoDB"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecCtx(ctx, "INSERT INTO "+table+" (id, value) VALUES (1, 0), (2, 0)"); err != nil {
		t.Fatal(err)
	}

	var attempts [2]atomic.Int32
	var firstLocks sync.WaitGroup
	firstLocks.Add(2)
	releaseSecondLock := make(chan struct{})
	run := func(worker, firstID, secondID int) error {
		return TransactWithDeadlockRetry(ctx, db, func(txCtx context.Context, session sqlx.Session) error {
			attempt := attempts[worker].Add(1)
			conn := sqlx.NewSqlConnFromSession(session)
			if _, err := conn.ExecCtx(txCtx, "UPDATE "+table+" SET value = value + 1 WHERE id = ?", firstID); err != nil {
				return err
			}
			if attempt == 1 {
				firstLocks.Done()
				select {
				case <-releaseSecondLock:
				case <-txCtx.Done():
					return txCtx.Err()
				}
			}
			_, err := conn.ExecCtx(txCtx, "UPDATE "+table+" SET value = value + 1 WHERE id = ?", secondID)
			return err
		})
	}

	errs := make(chan error, 2)
	go func() { errs <- run(0, 1, 2) }()
	go func() { errs <- run(1, 2, 1) }()
	firstLocks.Wait()
	close(releaseSecondLock)
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}

	var rows []struct {
		ID    int64 `db:"id"`
		Value int64 `db:"value"`
	}
	if err := db.QueryRowsCtx(ctx, &rows, "SELECT id, value FROM "+table+" ORDER BY id"); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Value != 2 || rows[1].Value != 2 {
		t.Fatalf("rolled-back work leaked or committed work was lost: %+v", rows)
	}
	if attempts[0].Load() == 1 && attempts[1].Load() == 1 {
		t.Fatalf("expected a real deadlock retry, attempts=(%d,%d)", attempts[0].Load(), attempts[1].Load())
	}
	if attempts[0].Load() > contractTransactionMaxAttempts || attempts[1].Load() > contractTransactionMaxAttempts {
		t.Fatalf("retry bound exceeded: attempts=(%d,%d)", attempts[0].Load(), attempts[1].Load())
	}
}
