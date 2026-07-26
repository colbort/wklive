package tasks

import (
	"context"
	"time"

	"wklive/services/option/internal/config"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const minimumMarketSnapshotInboxRetention = 14 * 24 * time.Hour

type marketSnapshotInboxCleanupSettings struct {
	retention  time.Duration
	interval   time.Duration
	batchSize  int64
	maxBatches int
}

func StartMarketSnapshotInboxCleanup(ctx context.Context, svcCtx *svc.ServiceContext) {
	settings, retentionAdjusted := resolveMarketSnapshotInboxCleanupSettings(svcCtx.Config)
	if retentionAdjusted {
		logx.Errorf("MarketSnapshotInboxCleanup.RetentionHours is below the safe minimum; using %s", settings.retention)
	}

	go func() {
		ticker := time.NewTicker(settings.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				cutoff := now.Add(-settings.retention).Unix()
				var total int64
				for i := 0; i < settings.maxBatches; i++ {
					deleted, err := svcCtx.OptionMarketSnapshotInboxModel.DeleteBefore(ctx, cutoff, settings.batchSize)
					if err != nil {
						if ctx.Err() == nil {
							logx.Errorf("option market snapshot inbox cleanup failed: %v", err)
						}
						break
					}
					total += deleted
					if deleted < settings.batchSize {
						break
					}
				}
				if total > 0 {
					logx.Infof("option market snapshot inbox cleanup completed deleted=%d cutoff=%d", total, cutoff)
				}
			}
		}
	}()
}

func resolveMarketSnapshotInboxCleanupSettings(c config.Config) (marketSnapshotInboxCleanupSettings, bool) {
	raw := c.MarketSnapshotInboxCleanup
	retention := time.Duration(raw.RetentionHours) * time.Hour
	adjusted := false
	if retention <= 0 {
		retention = 30 * 24 * time.Hour
	} else if retention < minimumMarketSnapshotInboxRetention {
		retention = minimumMarketSnapshotInboxRetention
		adjusted = true
	}
	interval := time.Duration(raw.IntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = time.Hour
	}
	batchSize := raw.BatchSize
	if batchSize <= 0 || batchSize > 10000 {
		batchSize = 5000
	}
	maxBatches := raw.MaxBatchesPerRun
	if maxBatches <= 0 {
		maxBatches = 10
	}
	return marketSnapshotInboxCleanupSettings{
		retention:  retention,
		interval:   interval,
		batchSize:  batchSize,
		maxBatches: maxBatches,
	}, adjusted
}
