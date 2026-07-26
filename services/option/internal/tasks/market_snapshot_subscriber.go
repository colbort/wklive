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
	messageSuccess       atomic.Int64
	handlerAttemptFailed atomic.Int64
	updated              atomic.Int64
	duplicates           atomic.Int64
	restarts             atomic.Int64
}

func StartMarketSnapshotSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	var stats marketSnapshotConsumerStats
	go logMarketSnapshotConsumerStats(ctx, &stats)
	go func() {
		backoff := time.Second
		for ctx.Err() == nil {
			err := svcCtx.MarketSnapshotSubscriber.Subscribe(ctx, market.AuthoritativeSnapshotTopic, func(ctx context.Context, msg mq.Message) error {
				var event market.AuthoritativeSnapshotEvent
				if err := mq.Decode(msg, &event); err != nil {
					stats.handlerAttemptFailed.Add(1)
					return err
				}
				result, err := optionlogic.NewSyncMarketQuoteLogic(ctx, svcCtx).SyncAuthoritativeSnapshot(event)
				if err != nil {
					stats.handlerAttemptFailed.Add(1)
					return err
				}
				stats.messageSuccess.Add(1)
				stats.updated.Add(result.Updated)
				stats.duplicates.Add(result.Duplicates)
				return nil
			})
			if ctx.Err() != nil {
				return
			}
			stats.restarts.Add(1)
			logx.Errorf("option market snapshot subscriber stopped, restarting in %s: %v", backoff, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < 30*time.Second {
				backoff *= 2
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
			}
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
			logx.Infof("option market snapshot consumer metrics message_success=%d handler_attempt_failed=%d updated=%d duplicates=%d restarts=%d",
				stats.messageSuccess.Load(), stats.handlerAttemptFailed.Load(), stats.updated.Load(), stats.duplicates.Load(), stats.restarts.Load())
		}
	}
}
