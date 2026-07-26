package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	market "wklive/common/market"
	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type snapshotOutboxPayload struct {
	Snapshot *market.SettlementSnapshot `json:"snapshot"`
	Message  types.ClientMessage        `json:"message"`
	Quote    *types.QuotePayload        `json:"quote"`
}

func StartSnapshotOutbox(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		lastHealthCheck := time.Time{}
		for {
			if err := processSnapshotOutbox(ctx, svcCtx); err != nil && ctx.Err() == nil {
				logx.Errorf("snapshot outbox worker failed: %v", err)
			}
			if lastHealthCheck.IsZero() || time.Since(lastHealthCheck) >= 30*time.Second {
				checkSnapshotOutboxHealth(ctx, svcCtx, time.Now())
				lastHealthCheck = time.Now()
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func checkSnapshotOutboxHealth(ctx context.Context, svcCtx *svc.ServiceContext, now time.Time) {
	health, err := svcCtx.SnapshotOutboxModel.Health(ctx)
	if err != nil {
		if ctx.Err() == nil {
			logx.Errorf("snapshot outbox health query failed: %v", err)
		}
		return
	}
	oldestAge := int64(0)
	if health.OldestOpenAt > 0 {
		oldestAge = now.UnixMilli() - health.OldestOpenAt
	}
	if !snapshotOutboxUnhealthy(health, oldestAge) {
		return
	}
	logx.Errorf("snapshot outbox unhealthy pending=%d processing=%d failed=%d manual=%d oldest_open_age_ms=%d",
		health.PendingCount, health.ProcessingCount, health.FailedCount, health.ManualCount, oldestAge)
}

func snapshotOutboxUnhealthy(health *models.SnapshotOutboxHealth, oldestAgeMillis int64) bool {
	if health == nil {
		return true
	}
	return health.FailedCount > 0 || health.ManualCount > 0 ||
		(health.PendingCount+health.ProcessingCount > 0 && oldestAgeMillis > int64(time.Minute/time.Millisecond))
}

func processSnapshotOutbox(ctx context.Context, svcCtx *svc.ServiceContext) error {
	now := time.Now().UnixMilli()
	rows, err := svcCtx.SnapshotOutboxModel.FindPending(ctx, now, 100)
	if err != nil {
		return err
	}
	for _, row := range rows {
		claimed, claimErr := svcCtx.SnapshotOutboxModel.Claim(ctx, row.Id, now)
		if claimErr != nil {
			return claimErr
		}
		if !claimed {
			continue
		}
		if publishErr := publishSnapshotOutbox(ctx, svcCtx, row); publishErr != nil {
			if markErr := svcCtx.SnapshotOutboxModel.MarkFailure(ctx, row.Id, publishErr.Error(), time.Now().UnixMilli()); markErr != nil {
				return markErr
			}
			continue
		}
		if err = svcCtx.SnapshotOutboxModel.MarkSuccess(ctx, row.Id, time.Now().UnixMilli()); err != nil {
			return err
		}
	}
	return nil
}

func publishSnapshotOutbox(ctx context.Context, svcCtx *svc.ServiceContext, row *models.TItickSnapshotOutbox) error {
	var payload snapshotOutboxPayload
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		return err
	}
	if payload.Snapshot == nil {
		return fmt.Errorf("outbox %s has no snapshot", row.SnapshotId)
	}
	if row.RedisPublishedAt == 0 {
		if err := svcCtx.MarketDataCache.PublishAuthoritativeSnapshot(ctx, payload.Snapshot); err != nil {
			return err
		}
		if err := svcCtx.SnapshotOutboxModel.MarkRedisPublished(ctx, row.Id, time.Now().UnixMilli()); err != nil {
			return fmt.Errorf("checkpoint Redis snapshot publication: %w", err)
		}
		row.RedisPublishedAt = time.Now().UnixMilli()
	}
	// Migrated repair rows intentionally contain only the snapshot. Redis repair
	// is complete even though the original full quote is no longer available.
	if payload.Quote == nil {
		if row.EventPublishedAt == 0 {
			if err := svcCtx.SnapshotOutboxModel.MarkEventPublished(ctx, row.Id, time.Now().UnixMilli()); err != nil {
				return fmt.Errorf("checkpoint skipped market event publication: %w", err)
			}
		}
		return nil
	}
	if row.EventPublishedAt > 0 {
		return nil
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
		return err
	}
	if err := svcCtx.SnapshotOutboxModel.MarkEventPublished(ctx, row.Id, time.Now().UnixMilli()); err != nil {
		return fmt.Errorf("checkpoint market event publication: %w", err)
	}
	return nil
}
