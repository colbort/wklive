// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"
	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarketSnapshotListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMarketSnapshotListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarketSnapshotListLogic {
	return &GetMarketSnapshotListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMarketSnapshotListLogic) GetMarketSnapshotList(req *types.GetMarketSnapshotListReq) (resp *types.GetMarketSnapshotListResp, err error) {
	return logicutil.Proxy[types.GetMarketSnapshotListResp](l.ctx, req, l.svcCtx.TradeCli.GetMarketSnapshotList)
}
