// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListMarketSnapshotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListMarketSnapshotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListMarketSnapshotsLogic {
	return &AdminListMarketSnapshotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListMarketSnapshotsLogic) AdminListMarketSnapshots(req *types.ListMarketSnapshotsReq) (resp *types.ListMarketSnapshotsResp, err error) {
	return logicutil.Proxy[types.ListMarketSnapshotsResp](l.ctx, req, l.svcCtx.OptionCli.AdminListMarketSnapshots)
}
