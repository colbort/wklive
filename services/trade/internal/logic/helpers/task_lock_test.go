package helpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"wklive/proto/trade"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestTaskLockRenewalFailureCancelsWork(t *testing.T) {
	renewFailure := errors.New("redis unavailable")
	var refreshes atomic.Int64
	workCanceled := make(chan struct{})

	resp, err := runWithTaskLockRenewal(
		context.Background(),
		time.Millisecond,
		func(context.Context) error {
			refreshes.Add(1)
			return renewFailure
		},
		func(ctx context.Context) (*trade.TradeTaskResp, error) {
			<-ctx.Done()
			close(workCanceled)
			return nil, ctx.Err()
		},
	)
	if resp != nil {
		t.Fatalf("renewal failure returned a response: %+v", resp)
	}
	if err == nil || !strings.Contains(err.Error(), renewFailure.Error()) {
		t.Fatalf("renewal failure error=%v", err)
	}
	if refreshes.Load() != 1 {
		t.Fatalf("refresh count=%d want=1", refreshes.Load())
	}
	select {
	case <-workCanceled:
	default:
		t.Fatal("task context was not canceled after lock renewal failure")
	}
}

func TestTaskLockRenewalReturnsTaskResultWhileLeaseIsOwned(t *testing.T) {
	want := OkTaskResp()
	resp, err := runWithTaskLockRenewal(
		context.Background(),
		time.Hour,
		func(context.Context) error { return nil },
		func(context.Context) (*trade.TradeTaskResp, error) { return want, nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if resp != want {
		t.Fatalf("response=%p want=%p", resp, want)
	}
}

func TestRedisEvalOKRequiresPositiveLeaseResult(t *testing.T) {
	for _, value := range []any{int64(1), int(2), int32(3), uint64(4)} {
		if !RedisEvalOK(value) {
			t.Fatalf("positive lease result rejected: %#v", value)
		}
	}
	for _, value := range []any{int64(0), int(-1), "1", nil} {
		if RedisEvalOK(value) {
			t.Fatalf("invalid lease result accepted: %#v", value)
		}
	}
}

func TestTaskLockOwnershipAgainstRedis(t *testing.T) {
	endpoint := os.Getenv("TRADE_REDIS_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("TRADE_REDIS_TEST_ENDPOINT is not set")
	}
	rds, err := redis.NewRedis(redis.RedisConf{Host: endpoint, Type: redis.NodeType})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	key := fmt.Sprintf("trade:acceptance:task-lock:%d", time.Now().UnixNano())
	ownerA, ownerB := "acceptance-worker-a", "acceptance-worker-b"
	t.Cleanup(func() {
		_ = ReleaseTaskLock(context.Background(), rds, key, ownerA)
		_ = ReleaseTaskLock(context.Background(), rds, key, ownerB)
	})

	if err = AcquireTaskLock(ctx, rds, key, ownerA); err != nil {
		t.Fatal(err)
	}
	if err = AcquireTaskLock(ctx, rds, key, ownerB); err == nil {
		t.Fatal("second worker acquired an active task lease")
	}
	if err = RefreshTaskLock(ctx, rds, key, ownerB); err == nil {
		t.Fatal("non-owner refreshed another worker's task lease")
	}
	if err = ReleaseTaskLock(ctx, rds, key, ownerB); err != nil {
		t.Fatal(err)
	}
	if err = AcquireTaskLock(ctx, rds, key, ownerB); err == nil {
		t.Fatal("non-owner release removed the active task lease")
	}
	if err = RefreshTaskLock(ctx, rds, key, ownerA); err != nil {
		t.Fatal(err)
	}
	if err = ReleaseTaskLock(ctx, rds, key, ownerA); err != nil {
		t.Fatal(err)
	}
	if err = AcquireTaskLock(ctx, rds, key, ownerB); err != nil {
		t.Fatalf("new worker could not acquire released task lease: %v", err)
	}
}
