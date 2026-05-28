package logic

import (
	"context"
	"time"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CleanMarketSnapshotsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCleanMarketSnapshotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CleanMarketSnapshotsLogic {
	return &CleanMarketSnapshotsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 期权行情快照归档/清理
func (l *CleanMarketSnapshotsLogic) CleanMarketSnapshots(in *option.OptionTaskReq) (*option.OptionTaskResp, error) {
	return runOptionTaskWithLock(l.ctx, l.svcCtx, "clean_market_snapshots", func() (*option.OptionTaskResp, error) {
		cutoff := time.Now().AddDate(0, 0, -30).Unix()
		cursor := int64(0)
		for {
			items, _, err := l.svcCtx.OptionMarketSnapshotModel.FindPage(l.ctx, models.OptionMarketSnapshotPageFilter{
				SnapshotEnd: cutoff,
			}, cursor, 100)
			if err != nil {
				return nil, err
			}
			if len(items) == 0 {
				break
			}
			for _, item := range items {
				cursor = item.Id
				if err := l.svcCtx.OptionMarketSnapshotModel.Delete(l.ctx, item.Id); err != nil {
					return nil, err
				}
			}
			if len(items) < 100 {
				break
			}
		}
		return okOptionTaskResp(), nil
	})
}
