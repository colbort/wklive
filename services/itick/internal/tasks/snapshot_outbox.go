package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	market "wklive/common/market"
	"wklive/proto/option"
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
		for {
			if err := processSnapshotOutbox(ctx, svcCtx); err != nil && ctx.Err() == nil {
				logx.Errorf("snapshot outbox worker failed: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
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
	if err := svcCtx.MarketDataCache.PublishAuthoritativeSnapshot(ctx, payload.Snapshot); err != nil {
		return err
	}
	// Migrated repair rows intentionally contain only the snapshot. Redis repair
	// is complete even though the original full quote is no longer available.
	if payload.Quote == nil {
		return nil
	}
	resp, err := svcCtx.OptionCli.SyncMarketQuote(ctx, &option.SyncMarketQuoteReq{CategoryCode: payload.Message.CategoryCode, Market: payload.Message.Market, Symbol: payload.Message.Symbol, UnderlyingPrice: payload.Quote.LastPriceText, OpenPrice: strconv.FormatFloat(payload.Quote.Open, 'f', -1, 64), HighPrice: strconv.FormatFloat(payload.Quote.High, 'f', -1, 64), LowPrice: strconv.FormatFloat(payload.Quote.Low, 'f', -1, 64), Volume: strconv.FormatFloat(payload.Quote.Volume, 'f', -1, 64), Turnover: strconv.FormatFloat(payload.Quote.Turnover, 'f', -1, 64), QuoteTs: payload.Quote.Ts})
	if err != nil {
		return err
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("option quote sync rejected")
	}
	return nil
}
