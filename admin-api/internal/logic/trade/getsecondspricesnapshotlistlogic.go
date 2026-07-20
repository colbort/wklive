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

type GetSecondsPriceSnapshotListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSecondsPriceSnapshotListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSecondsPriceSnapshotListLogic {
	return &GetSecondsPriceSnapshotListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSecondsPriceSnapshotListLogic) GetSecondsPriceSnapshotList(req *types.GetSecondsPriceSnapshotListReq) (resp *types.GetSecondsPriceSnapshotListResp, err error) {
	return logicutil.Proxy[types.GetSecondsPriceSnapshotListResp](l.ctx, req, l.svcCtx.TradeCli.GetSecondsPriceSnapshotList)
}
