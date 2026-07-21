package tasks

import (
	"context"
	"time"

	market "wklive/common/market"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// StartAuthoritativeSnapshotRebuild reconstructs the disposable Redis index
// from the permanent MySQL archive on every service start. It deliberately
// bypasses outbox delivery state, so a Redis loss is recoverable even when all
// historical outbox rows were already marked successful.
func StartAuthoritativeSnapshotRebuild(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		var afterID int64
		for ctx.Err() == nil {
			rows, err := svcCtx.AuthoritativeSnapshotModel.FindAfterID(ctx, afterID, 500)
			if err != nil {
				logx.Errorf("authoritative snapshot rebuild scan failed: %v", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
					continue
				}
			}
			if len(rows) == 0 {
				return
			}
			for _, row := range rows {
				snapshot := &market.SettlementSnapshot{SnapshotID: row.SnapshotId, Authority: row.Authority, Kind: row.SnapshotKind, CategoryCode: row.CategoryCode, Market: row.Market, Symbol: row.Symbol, Price: row.Price.String(), Source: row.Authority, SourceTimestamp: row.SourceTimestamp, SnapshotTimestamp: row.SnapshotTimestamp, Revision: row.Revision, FormulaVersion: row.FormulaVersion, Confirmed: true}
				if err = svcCtx.MarketDataCache.PublishAuthoritativeSnapshot(ctx, snapshot); err != nil {
					logx.Errorf("authoritative snapshot rebuild publish failed, snapshotId=%s err=%v", row.SnapshotId, err)
					break
				}
				afterID = row.Id
			}
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
			}
		}
	}()
}
