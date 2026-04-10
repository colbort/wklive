package logic

import (
	"context"
	"errors"

	pageutil "wklive/common/pageutil"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListMarketSnapshotsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListMarketSnapshotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListMarketSnapshotsLogic {
	return &AdminListMarketSnapshotsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询期权行情快照列表
func (l *AdminListMarketSnapshotsLogic) AdminListMarketSnapshots(in *option.ListMarketSnapshotsReq) (*option.ListMarketSnapshotsResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionMarketSnapshotModel.FindPage(l.ctx, models.OptionMarketSnapshotPageFilter{
		TenantId:      in.TenantId,
		ContractId:    in.ContractId,
		SnapshotStart: pageutil.TimeRangeStart(in.TimeRange),
		SnapshotEnd:   pageutil.TimeRangeEnd(in.TimeRange),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	list := make([]*option.OptionMarketSnapshot, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		list = append(list, toMarketSnapshotProto(item))
	}

	return &option.ListMarketSnapshotsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}
