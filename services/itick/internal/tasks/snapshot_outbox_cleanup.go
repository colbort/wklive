package tasks

import (
	"context"
	"time"

	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultSnapshotOutboxSuccessRetention  = 30 * time.Minute
	defaultSnapshotOutboxCleanupInterval   = time.Minute
	defaultSnapshotOutboxCleanupBatchSize  = int64(5000)
	defaultSnapshotOutboxCleanupMaxBatches = 10
	defaultSnapshotOutboxCleanupBatchPause = 100 * time.Millisecond
)

type snapshotOutboxCleanupSettings struct {
	retention  time.Duration
	interval   time.Duration
	batchSize  int64
	maxBatches int
	batchPause time.Duration
}

func snapshotOutboxCleanupSettingsFrom(svcCtx *svc.ServiceContext) snapshotOutboxCleanupSettings {
	c := svcCtx.Config.SnapshotOutboxCleanup
	s := snapshotOutboxCleanupSettings{
		retention:  time.Duration(c.SuccessRetentionMinutes) * time.Minute,
		interval:   time.Duration(c.IntervalSeconds) * time.Second,
		batchSize:  c.BatchSize,
		maxBatches: c.MaxBatchesPerRun,
		batchPause: time.Duration(c.BatchPauseMs) * time.Millisecond,
	}
	if s.retention <= 0 {
		s.retention = defaultSnapshotOutboxSuccessRetention
	}
	if s.interval <= 0 {
		s.interval = defaultSnapshotOutboxCleanupInterval
	}
	if s.batchSize <= 0 || s.batchSize > 10000 {
		s.batchSize = defaultSnapshotOutboxCleanupBatchSize
	}
	if s.maxBatches <= 0 {
		s.maxBatches = defaultSnapshotOutboxCleanupMaxBatches
	}
	if s.batchPause <= 0 {
		s.batchPause = defaultSnapshotOutboxCleanupBatchPause
	}
	return s
}

func StartSnapshotOutboxCleanup(ctx context.Context, svcCtx *svc.ServiceContext) {
	settings := snapshotOutboxCleanupSettingsFrom(svcCtx)
	go func() {
		ticker := time.NewTicker(settings.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cleanupSnapshotOutbox(ctx, svcCtx.SnapshotOutboxModel.DeleteSucceededBefore, settings, time.Now())
			}
		}
	}()
}

func cleanupSnapshotOutbox(
	ctx context.Context,
	deleteSucceededBefore func(context.Context, int64, int64) (int64, error),
	settings snapshotOutboxCleanupSettings,
	now time.Time,
) {
	cutoff := now.Add(-settings.retention).UnixMilli()
	var total int64
	for batch := 0; batch < settings.maxBatches; batch++ {
		deleted, err := deleteSucceededBefore(ctx, cutoff, settings.batchSize)
		if err != nil {
			if ctx.Err() == nil {
				logx.Errorf("cleanup snapshot outbox failed: %v", err)
			}
			return
		}
		total += deleted
		if deleted < settings.batchSize {
			break
		}
		if settings.batchPause > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(settings.batchPause):
			}
		}
	}
	if total > 0 {
		logx.Infof("snapshot outbox cleanup completed deleted=%d cutoff=%d", total, cutoff)
	}
}
