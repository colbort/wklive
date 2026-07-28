package helpers

import (
	"context"
	"errors"
	"testing"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestDeadlockTransactionRetriesThenSucceeds(t *testing.T) {
	attempts := 0
	runner := func(context.Context, func(context.Context, sqlx.Session) error) error {
		attempts++
		if attempts < 3 {
			return &mysql.MySQLError{Number: 1213, Message: "deadlock"}
		}
		return nil
	}
	if err := transactWithDeadlockRetry(context.Background(), 3, 0, runner, func(context.Context, sqlx.Session) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("attempts=%d want=3", attempts)
	}
}

func TestLockWaitTimeoutIsBounded(t *testing.T) {
	attempts := 0
	runner := func(context.Context, func(context.Context, sqlx.Session) error) error {
		attempts++
		return &mysql.MySQLError{Number: 1205, Message: "lock wait timeout"}
	}
	err := transactWithDeadlockRetry(context.Background(), 3, 0, runner, func(context.Context, sqlx.Session) error {
		return nil
	})
	if err == nil || attempts != 3 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
}

func TestBusinessTransactionErrorIsNotRetried(t *testing.T) {
	want := errors.New("position changed")
	attempts := 0
	runner := func(context.Context, func(context.Context, sqlx.Session) error) error {
		attempts++
		return want
	}
	err := transactWithDeadlockRetry(context.Background(), 3, time.Millisecond, runner, func(context.Context, sqlx.Session) error {
		return nil
	})
	if !errors.Is(err, want) || attempts != 1 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
}
