package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

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
	// todo: add your logic here and delete this line

	return &option.ListMarketSnapshotsResp{}, nil
}
