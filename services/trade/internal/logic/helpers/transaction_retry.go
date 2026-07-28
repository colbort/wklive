package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	contractTransactionMaxAttempts = 3
	contractTransactionRetryBase   = 10 * time.Millisecond
)

type transactionRunner func(context.Context, func(context.Context, sqlx.Session) error) error

// TransactWithDeadlockRetry retries only database concurrency failures whose
// transaction has been rolled back by TransactCtx. Business validation,
// duplicate facts and external RPC errors are never replayed here.
func TransactWithDeadlockRetry(
	ctx context.Context,
	db sqlx.SqlConn,
	fn func(context.Context, sqlx.Session) error,
) error {
	if db == nil {
		return errors.New("transaction database is nil")
	}
	return transactWithDeadlockRetry(ctx, contractTransactionMaxAttempts, contractTransactionRetryBase, db.TransactCtx, fn)
}

func transactWithDeadlockRetry(
	ctx context.Context,
	maxAttempts int,
	baseDelay time.Duration,
	transact transactionRunner,
	fn func(context.Context, sqlx.Session) error,
) error {
	if maxAttempts <= 0 || baseDelay < 0 || transact == nil || fn == nil {
		return errors.New("invalid deadlock retry configuration")
	}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = transact(ctx, fn)
		if lastErr == nil || !isRetryableMySQLTransactionError(lastErr) {
			return lastErr
		}
		if attempt == maxAttempts {
			break
		}
		timer := time.NewTimer(baseDelay * time.Duration(attempt))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf("mysql transaction concurrency failure after %d attempts: %w", maxAttempts, lastErr)
}

func isRetryableMySQLTransactionError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		return false
	}
	return mysqlErr.Number == 1213 || mysqlErr.Number == 1205
}
