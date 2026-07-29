package tasks

import (
	"context"
	"time"

	"wklive/services/market/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// StartLegacyAuthoritativeCacheCleanup removes the obsolete unbounded v1/v2
// cache layout with SCAN + UNLINK. It is disabled by default for rolling-deploy safety.
func StartLegacyAuthoritativeCacheCleanup(ctx context.Context, svcCtx *svc.ServiceContext) {
	c := svcCtx.Config.AuthoritativeCache
	if !c.LegacyCleanupEnabled {
		return
	}
	count := c.LegacyCleanupScanCount
	if count <= 0 || count > 5000 {
		count = 500
	}
	interval := time.Duration(c.LegacyCleanupIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = time.Second
	}
	go func() {
		var cursor uint64
		var total int64
		for ctx.Err() == nil {
			next, deleted, err := svcCtx.MarketDataCache.CleanupLegacyAuthoritativeCache(ctx, cursor, count)
			if err != nil {
				if ctx.Err() == nil {
					logx.Errorf("legacy authoritative cache cleanup failed: %v", err)
				}
				return
			}
			total += deleted
			cursor = next
			if cursor == 0 {
				logx.Infof("legacy authoritative cache cleanup completed deleted=%d", total)
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
			}
		}
	}()
}
