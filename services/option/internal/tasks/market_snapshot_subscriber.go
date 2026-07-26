package tasks

import (
	"context"
	"sync/atomic"
	"time"

	market "wklive/common/market"
	mq "wklive/common/mq/kafka"
	optionlogic "wklive/services/option/internal/logic/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type marketSnapshotConsumerStats struct {
	success    atomic.Int64
	failed     atomic.Int64
	updated    atomic.Int64
	duplicates atomic.Int64
}

func StartMarketSnapshotSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	var stats marketSnapshotConsumerStats
	go logMarketSnapshotConsumerStats(ctx, &stats)
	go func() {
		err := svcCtx.MarketSnapshotSubscriber.Subscribe(ctx, market.AuthoritativeSnapshotTopic, func(ctx context.Context, msg mq.Message) error {
			var event market.AuthoritativeSnapshotEvent
			if err := mq.Decode(msg, &event); err != nil {
				stats.failed.Add(1)
				return err
			}
			result, err := optionlogic.NewSyncMarketQuoteLogic(ctx, svcCtx).SyncAuthoritativeSnapshot(event)
			if err != nil {
				stats.failed.Add(1)
				return err
			}
			stats.success.Add(1)
			stats.updated.Add(result.Updated)
			stats.duplicates.Add(result.Duplicates)
			return nil
		})
		if err != nil && ctx.Err() == nil {
			logx.Errorf("option market snapshot subscriber stopped: %v", err)
		}
	}()
}

func logMarketSnapshotConsumerStats(ctx context.Context, stats *marketSnapshotConsumerStats) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logx.Infof("option market snapshot consumer metrics success=%d failed=%d updated=%d duplicates=%d",
				stats.success.Load(), stats.failed.Load(), stats.updated.Load(), stats.duplicates.Load())
		}
	}
}
