package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	market "wklive/common/market"
	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultSnapshotOutboxWorkerCount    = 32
	defaultSnapshotOutboxBatchSize      = int64(512)
	defaultSnapshotOutboxIdleInterval   = 100 * time.Millisecond
	defaultSnapshotOutboxHealthInterval = 30 * time.Second
)

type snapshotOutboxPayload struct {
	Snapshot *market.SettlementSnapshot `json:"snapshot"`
	Message  types.ClientMessage        `json:"message"`
	Quote    *types.QuotePayload        `json:"quote"`
}

func StartSnapshotOutbox(ctx context.Context, svcCtx *svc.ServiceContext) {
	workerCount := svcCtx.Config.SnapshotOutbox.WorkerCount
	if workerCount <= 0 || workerCount > 256 {
		workerCount = defaultSnapshotOutboxWorkerCount
	}
	batchSize := svcCtx.Config.SnapshotOutbox.BatchSize
	if batchSize < int64(workerCount) || batchSize > 5000 {
		batchSize = defaultSnapshotOutboxBatchSize
	}
	idleInterval := time.Duration(svcCtx.Config.SnapshotOutbox.IdleIntervalMs) * time.Millisecond
	if idleInterval <= 0 {
		idleInterval = defaultSnapshotOutboxIdleInterval
	}

	// 健康检查独立运行，不能再被一个耗时的发布批次阻塞。
	go func() {
		ticker := time.NewTicker(defaultSnapshotOutboxHealthInterval)
		defer ticker.Stop()
		var previous *snapshotOutboxHealthSample
		previous = checkSnapshotOutboxHealth(ctx, svcCtx, time.Now(), previous)
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				previous = checkSnapshotOutboxHealth(ctx, svcCtx, now, previous)
			}
		}
	}()

	go func() {
		for {
			processed, err := processSnapshotOutbox(ctx, svcCtx, workerCount, batchSize)
			if err != nil && ctx.Err() == nil {
				logx.Errorf("snapshot outbox worker failed: %v", err)
			}
			// 有满批积压时立即继续，不再固定睡眠一秒。
			if processed >= batchSize {
				continue
			}
			timer := time.NewTimer(idleInterval)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return
			case <-timer.C:
			}
		}
	}()
}

type snapshotOutboxHealthSample struct {
	at       time.Time
	openRows int64
}

func checkSnapshotOutboxHealth(ctx context.Context, svcCtx *svc.ServiceContext, now time.Time, previous *snapshotOutboxHealthSample) *snapshotOutboxHealthSample {
	health, err := svcCtx.SnapshotOutboxModel.Health(ctx)
	if err != nil {
		if ctx.Err() == nil {
			logx.Errorf("snapshot outbox health query failed: %v", err)
		}
		return previous
	}
	current := &snapshotOutboxHealthSample{at: now, openRows: health.PendingCount + health.ProcessingCount}
	oldestAge := int64(0)
	if health.OldestOpenAt > 0 {
		oldestAge = now.UnixMilli() - health.OldestOpenAt
	}
	if !snapshotOutboxUnhealthy(health, oldestAge) {
		return current
	}
	drainPerSecond, etaSeconds := snapshotOutboxDrainMetrics(previous, current)
	logx.Errorf("snapshot outbox unhealthy pending=%d processing=%d failed=%d manual=%d oldest_open_age_ms=%d drain_per_sec=%.2f eta_seconds=%d",
		health.PendingCount, health.ProcessingCount, health.FailedCount, health.ManualCount, oldestAge, drainPerSecond, etaSeconds)
	return current
}

func snapshotOutboxDrainMetrics(previous, current *snapshotOutboxHealthSample) (float64, int64) {
	if previous == nil || current == nil || !current.at.After(previous.at) {
		return 0, -1
	}
	elapsed := current.at.Sub(previous.at).Seconds()
	drained := previous.openRows - current.openRows
	if elapsed <= 0 || drained <= 0 {
		return 0, -1
	}
	rate := float64(drained) / elapsed
	return rate, int64(float64(current.openRows) / rate)
}

func snapshotOutboxUnhealthy(health *models.SnapshotOutboxHealth, oldestAgeMillis int64) bool {
	if health == nil {
		return true
	}
	return health.FailedCount > 0 || health.ManualCount > 0 ||
		(health.PendingCount+health.ProcessingCount > 0 && oldestAgeMillis > int64(time.Minute/time.Millisecond))
}

func processSnapshotOutbox(ctx context.Context, svcCtx *svc.ServiceContext, workerCount int, batchSize int64) (int64, error) {
	now := time.Now().UnixMilli()
	rows, err := svcCtx.SnapshotOutboxModel.FindPending(ctx, now, batchSize)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	jobs := make(chan *models.TItickSnapshotOutbox)
	errs := make(chan error, len(rows))
	var claimedCount atomic.Int64
	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for row := range jobs {
				claimed, claimErr := svcCtx.SnapshotOutboxModel.Claim(ctx, row.Id, time.Now().UnixMilli())
				if claimErr != nil {
					errs <- claimErr
					continue
				}
				if !claimed {
					continue
				}
				claimedCount.Add(1)
				completed, publishErr := publishSnapshotOutbox(ctx, svcCtx, row)
				if publishErr != nil {
					if markErr := svcCtx.SnapshotOutboxModel.MarkFailure(ctx, row.Id, publishErr.Error(), time.Now().UnixMilli()); markErr != nil {
						errs <- markErr
					}
					continue
				}
				if !completed {
					if markErr := svcCtx.SnapshotOutboxModel.MarkSuccess(ctx, row.Id, time.Now().UnixMilli()); markErr != nil {
						errs <- markErr
					}
				}
			}
		}()
	}
	for _, row := range rows {
		select {
		case <-ctx.Done():
			close(jobs)
			workers.Wait()
			close(errs)
			return claimedCount.Load(), ctx.Err()
		case jobs <- row:
		}
	}
	close(jobs)
	workers.Wait()
	close(errs)
	for workerErr := range errs {
		return claimedCount.Load(), workerErr
	}
	return claimedCount.Load(), nil
}

// publishSnapshotOutbox returns completed=true when it atomically moved the
// row to success. completed=false means the caller must close a row whose
// publication checkpoints were already present from an earlier attempt.
func publishSnapshotOutbox(ctx context.Context, svcCtx *svc.ServiceContext, row *models.TItickSnapshotOutbox) (bool, error) {
	var payload snapshotOutboxPayload
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		return false, err
	}
	if payload.Snapshot == nil {
		return false, fmt.Errorf("outbox %s has no snapshot", row.SnapshotId)
	}
	if row.RedisPublishedAt == 0 {
		if err := svcCtx.MarketDataCache.PublishAuthoritativeSnapshot(ctx, payload.Snapshot); err != nil {
			return false, err
		}
		if err := svcCtx.SnapshotOutboxModel.MarkRedisPublished(ctx, row.Id, time.Now().UnixMilli()); err != nil {
			return false, fmt.Errorf("checkpoint Redis snapshot publication: %w", err)
		}
		row.RedisPublishedAt = time.Now().UnixMilli()
	}
	// Migrated repair rows intentionally contain only the snapshot. Redis repair
	// is complete even though the original full quote is no longer available.
	if payload.Quote == nil {
		if row.EventPublishedAt == 0 {
			if err := svcCtx.SnapshotOutboxModel.CompleteAfterEventPublished(ctx, row.Id, time.Now().UnixMilli()); err != nil {
				return false, fmt.Errorf("complete skipped market event publication: %w", err)
			}
			row.EventPublishedAt = time.Now().UnixMilli()
			return true, nil
		}
		return false, nil
	}
	if row.EventPublishedAt > 0 {
		return false, nil
	}
	event := market.AuthoritativeSnapshotEvent{
		Version:         market.AuthoritativeSnapshotEventVersion,
		EventID:         payload.Snapshot.SnapshotID,
		SnapshotID:      payload.Snapshot.SnapshotID,
		CategoryCode:    payload.Message.CategoryCode,
		Market:          payload.Message.Market,
		Symbol:          payload.Message.Symbol,
		UnderlyingPrice: payload.Quote.LastPriceText,
		OpenPrice:       strconv.FormatFloat(payload.Quote.Open, 'f', -1, 64),
		HighPrice:       strconv.FormatFloat(payload.Quote.High, 'f', -1, 64),
		LowPrice:        strconv.FormatFloat(payload.Quote.Low, 'f', -1, 64),
		Volume:          strconv.FormatFloat(payload.Quote.Volume, 'f', -1, 64),
		Turnover:        strconv.FormatFloat(payload.Quote.Turnover, 'f', -1, 64),
		QuoteTimestamp:  payload.Quote.Ts,
		PublishedAt:     time.Now().UnixMilli(),
	}
	if err := svcCtx.SnapshotPublisher.PublishKey(ctx, market.AuthoritativeSnapshotTopic, []byte(strings.ToUpper(event.PartitionKey())), event); err != nil {
		return false, err
	}
	if err := svcCtx.SnapshotOutboxModel.CompleteAfterEventPublished(ctx, row.Id, time.Now().UnixMilli()); err != nil {
		return false, fmt.Errorf("complete market event publication: %w", err)
	}
	row.EventPublishedAt = time.Now().UnixMilli()
	return true, nil
}
