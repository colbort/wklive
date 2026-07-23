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

type ListMarketSnapshotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMarketSnapshotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMarketSnapshotsLogic {
	return &ListMarketSnapshotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMarketSnapshotsLogic) ListMarketSnapshots(req *types.ListMarketSnapshotsReq) (resp *types.ListMarketSnapshotsResp, err error) {
	return logicutil.Proxy[types.ListMarketSnapshotsResp](l.ctx, req, l.svcCtx.OptionCli.ListMarketSnapshots)
}
