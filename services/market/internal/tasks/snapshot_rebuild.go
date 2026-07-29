package tasks

import (
	"context"
	"errors"
	"math"

	market "wklive/common/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

// StartAuthoritativeSnapshotRebuild reconstructs only the latest snapshot for
// each product. Historical reads fall back to the permanent MySQL archive, so
// service startup never republishes the complete archive into Redis.
func StartAuthoritativeSnapshotRebuild(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		keys, err := svcCtx.AuthoritativeSnapshotModel.FindProductKeys(ctx)
		if err != nil {
			logx.Errorf("authoritative snapshot rebuild key scan failed: %v", err)
			return
		}
		for _, key := range keys {
			if ctx.Err() != nil {
				return
			}
			row, findErr := svcCtx.AuthoritativeSnapshotModel.FindAtOrBefore(
				ctx,
				key.Authority,
				key.SnapshotKind,
				key.CategoryCode,
				key.Market,
				key.Symbol,
				math.MaxInt64,
				0,
			)
			if errors.Is(findErr, models.ErrNotFound) {
				continue
			}
			if findErr != nil {
				logx.Errorf("authoritative snapshot rebuild latest lookup failed, authority=%s kind=%s category=%s market=%s symbol=%s err=%v", key.Authority, key.SnapshotKind, key.CategoryCode, key.Market, key.Symbol, findErr)
				continue
			}
			snapshot := &market.SettlementSnapshot{SnapshotID: row.SnapshotId, Authority: row.Authority, Kind: row.SnapshotKind, CategoryCode: row.CategoryCode, Market: row.Market, Symbol: row.Symbol, Price: row.Price.String(), Source: row.Authority, SourceTimestamp: row.SourceTimestamp, SnapshotTimestamp: row.SnapshotTimestamp, Revision: row.Revision, FormulaVersion: row.FormulaVersion, Confirmed: true}
			if err = svcCtx.MarketDataCache.PublishAuthoritativeSnapshot(ctx, snapshot); err != nil {
				logx.Errorf("authoritative snapshot rebuild publish failed, snapshotId=%s err=%v", row.SnapshotId, err)
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
