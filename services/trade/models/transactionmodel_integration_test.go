package models

import (
	"context"
	"fmt"
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
		return transactWithRetry(ctx, transactionMaxAttempts, transactionRetryBase, db.TransactCtx, func(txCtx context.Context, session sqlx.Session) error {
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
	if attempts[0].Load() > transactionMaxAttempts || attempts[1].Load() > transactionMaxAttempts {
		t.Fatalf("retry bound exceeded: attempts=(%d,%d)", attempts[0].Load(), attempts[1].Load())
	}
}

// TestDeadlockRetryWithContractFactsAgainstMySQL proves that the retry wrapper
// also preserves the real contract facts touched by settlement workflows. It
// uses a closed Position and a successful Instruction so background workers
// cannot claim the isolated rows.
func TestDeadlockRetryWithContractFactsAgainstMySQL(t *testing.T) {
	dsn := os.Getenv("TRADE_MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("TRADE_MYSQL_TEST_DSN is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := sqlx.NewMysql(dsn)
	unique := time.Now().UnixNano()
	tenantID := int64(910000000 + unique%10000000)
	userID := tenantID
	symbolID := int64(920000000 + unique%10000000)
	instructionNo := fmt.Sprintf("DEADLOCK-CONTRACT-FACTS-%d", unique)
	now := time.Now().UnixMilli()

	positionResult, err := db.ExecCtx(ctx, `
		INSERT INTO t_contract_position (
			tenant_id, user_id, symbol_id, contract_type, contract_value_type,
			position_side, margin_mode, status, leverage, qty, avail_qty,
			frozen_qty, margin_asset, version, create_times, update_times
		) VALUES (?, ?, ?, 1, 1, 2, 2, 5, 1, 0, 0, 0, 'USDT', 0, ?, ?)`,
		tenantID, userID, symbolID, now, now)
	if err != nil {
		t.Fatal(err)
	}
	positionID, err := positionResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	instructionResult, err := db.ExecCtx(ctx, `
		INSERT INTO t_trade_settlement_instruction (
			tenant_id, instruction_no, biz_type, biz_id, user_id, action,
			asset, amount, step_no, status, retry_count, create_times, update_times
		) VALUES (?, ?, 'deadlock-test', ?, ?, 6, 'USDT', 1, 1, 3, 0, ?, ?)`,
		tenantID, instructionNo, instructionNo, userID, now, now)
	if err != nil {
		_, _ = db.ExecCtx(context.Background(), "DELETE FROM t_contract_position WHERE id = ? AND tenant_id = ?", positionID, tenantID)
		t.Fatal(err)
	}
	instructionID, err := instructionResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = db.ExecCtx(cleanupCtx, "DELETE FROM t_trade_settlement_instruction WHERE id = ? AND tenant_id = ?", instructionID, tenantID)
		_, _ = db.ExecCtx(cleanupCtx, "DELETE FROM t_contract_position WHERE id = ? AND tenant_id = ?", positionID, tenantID)
	})

	var attempts [2]atomic.Int32
	var firstLocks sync.WaitGroup
	firstLocks.Add(2)
	releaseSecondLock := make(chan struct{})
	run := func(worker int, positionFirst bool) error {
		return transactWithRetry(ctx, transactionMaxAttempts, transactionRetryBase, db.TransactCtx, func(txCtx context.Context, session sqlx.Session) error {
			attempt := attempts[worker].Add(1)
			conn := sqlx.NewSqlConnFromSession(session)
			updatePosition := func() error {
				_, err := conn.ExecCtx(txCtx, `
					UPDATE t_contract_position
					SET realized_pnl = realized_pnl + 1, version = version + 1, update_times = ?
					WHERE id = ? AND tenant_id = ?`, now, positionID, tenantID)
				return err
			}
			updateInstruction := func() error {
				_, err := conn.ExecCtx(txCtx, `
					UPDATE t_trade_settlement_instruction
					SET retry_count = retry_count + 1, update_times = ?
					WHERE id = ? AND tenant_id = ?`, now, instructionID, tenantID)
				return err
			}

			var err error
			if positionFirst {
				err = updatePosition()
			} else {
				err = updateInstruction()
			}
			if err != nil {
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
			if positionFirst {
				return updateInstruction()
			}
			return updatePosition()
		})
	}

	errs := make(chan error, 2)
	go func() { errs <- run(0, true) }()
	go func() { errs <- run(1, false) }()
	firstLocks.Wait()
	close(releaseSecondLock)
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}

	var facts struct {
		RealizedPnl     string `db:"realized_pnl"`
		PositionVersion int64  `db:"position_version"`
		RetryCount      int64  `db:"retry_count"`
	}
	if err := db.QueryRowCtx(ctx, &facts, `
		SELECT p.realized_pnl, p.version AS position_version, i.retry_count
		FROM t_contract_position p
		JOIN t_trade_settlement_instruction i
		  ON i.id = ? AND i.tenant_id = p.tenant_id
		WHERE p.id = ? AND p.tenant_id = ?`,
		instructionID, positionID, tenantID); err != nil {
		t.Fatal(err)
	}
	if facts.RealizedPnl != "2.000000000000000000" || facts.PositionVersion != 2 || facts.RetryCount != 2 {
		t.Fatalf("contract facts lost or duplicated after deadlock retry: %+v", facts)
	}
	if attempts[0].Load() == 1 && attempts[1].Load() == 1 {
		t.Fatalf("expected a real deadlock retry, attempts=(%d,%d)", attempts[0].Load(), attempts[1].Load())
	}
	if attempts[0].Load() > transactionMaxAttempts || attempts[1].Load() > transactionMaxAttempts {
		t.Fatalf("retry bound exceeded: attempts=(%d,%d)", attempts[0].Load(), attempts[1].Load())
	}
	t.Logf(
		"real contract facts survived deadlock retry: tenant=%d position=%d instruction=%d attempts=(%d,%d)",
		tenantID, positionID, instructionID, attempts[0].Load(), attempts[1].Load(),
	)
}
