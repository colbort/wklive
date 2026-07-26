package tasks

import (
	"context"
	"time"

	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartMarketSnapshotInboxCleanup(ctx context.Context, svcCtx *svc.ServiceContext) {
	settings := svcCtx.Config.MarketSnapshotInboxCleanup
	retention := time.Duration(settings.RetentionHours) * time.Hour
	if retention <= 0 {
		retention = 30 * 24 * time.Hour
	}
	interval := time.Duration(settings.IntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = time.Hour
	}
	batchSize := settings.BatchSize
	if batchSize <= 0 || batchSize > 10000 {
		batchSize = 5000
	}
	maxBatches := settings.MaxBatchesPerRun
	if maxBatches <= 0 {
		maxBatches = 10
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				cutoff := now.Add(-retention).Unix()
				var total int64
				for i := 0; i < maxBatches; i++ {
					deleted, err := svcCtx.OptionMarketSnapshotInboxModel.DeleteBefore(ctx, cutoff, batchSize)
					if err != nil {
						if ctx.Err() == nil {
							logx.Errorf("option market snapshot inbox cleanup failed: %v", err)
						}
						break
					}
					total += deleted
					if deleted < batchSize {
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
