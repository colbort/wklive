package tasks

import (
	"context"
	"time"

	market "wklive/common/market"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// StartAuthoritativeSnapshotRebuild reconstructs only the latest snapshot for
// each product. Historical reads fall back to the permanent MySQL archive, so
// service startup never republishes the complete archive into Redis.
func StartAuthoritativeSnapshotRebuild(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		var afterID int64
		for ctx.Err() == nil {
			rows, err := svcCtx.AuthoritativeSnapshotModel.FindLatestPage(ctx, afterID, 500)
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
				break
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
		var revokeAfterID int64
		for ctx.Err() == nil {
			rows, err := svcCtx.SnapshotRevocationModel.FindAfterID(ctx, revokeAfterID, 500)
			if err != nil {
				logx.Errorf("authoritative snapshot revocation rebuild scan failed: %v", err)
				return
			}
			if len(rows) == 0 {
				return
			}
			for _, row := range rows {
				if err = svcCtx.MarketDataCache.RevokeAuthoritativeSnapshot(ctx, row.SnapshotId, row.ReplacementSnapshotId, row.Reason); err != nil {
					logx.Errorf("authoritative snapshot revocation rebuild failed, snapshotId=%s err=%v", row.SnapshotId, err)
					return
				}
				revokeAfterID = row.Id
			}
		}
	}()
}
