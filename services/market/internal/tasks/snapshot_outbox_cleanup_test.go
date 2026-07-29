package tasks

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCleanupSnapshotOutboxStopsAfterPartialBatch(t *testing.T) {
	now := time.UnixMilli(2_000_000)
	settings := snapshotOutboxCleanupSettings{
		retention:  30 * time.Minute,
		batchSize:  5000,
		maxBatches: 10,
	}
	var calls int
	cleanupSnapshotOutbox(context.Background(), func(_ context.Context, cutoff, limit int64) (int64, error) {
		calls++
		if want := now.Add(-settings.retention).UnixMilli(); cutoff != want {
			t.Fatalf("cutoff=%d want=%d", cutoff, want)
		}
		if limit != settings.batchSize {
			t.Fatalf("limit=%d want=%d", limit, settings.batchSize)
		}
		return 100, nil
	}, settings, now)
	if calls != 1 {
		t.Fatalf("calls=%d want=1", calls)
	}
}

func TestCleanupSnapshotOutboxHonorsMaximumBatches(t *testing.T) {
	settings := snapshotOutboxCleanupSettings{retention: time.Minute, batchSize: 10, maxBatches: 3}
	var calls int
	cleanupSnapshotOutbox(context.Background(), func(context.Context, int64, int64) (int64, error) {
		calls++
		return settings.batchSize, nil
	}, settings, time.Now())
	if calls != settings.maxBatches {
		t.Fatalf("calls=%d want=%d", calls, settings.maxBatches)
	}
}

func TestCleanupSnapshotOutboxStopsOnError(t *testing.T) {
	settings := snapshotOutboxCleanupSettings{retention: time.Minute, batchSize: 10, maxBatches: 3}
	var calls int
	cleanupSnapshotOutbox(context.Background(), func(context.Context, int64, int64) (int64, error) {
		calls++
		return 0, errors.New("delete failed")
	}, settings, time.Now())
	if calls != 1 {
		t.Fatalf("calls=%d want=1", calls)
	}
}
