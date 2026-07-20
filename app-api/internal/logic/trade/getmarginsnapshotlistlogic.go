// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarginSnapshotListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMarginSnapshotListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginSnapshotListLogic {
	return &GetMarginSnapshotListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMarginSnapshotListLogic) GetMarginSnapshotList(req *types.GetMarginSnapshotListReq) (resp *types.GetMarginSnapshotListResp, err error) {
	return logicutil.Proxy[types.GetMarginSnapshotListResp](l.ctx, req, l.svcCtx.TradeCli.GetMarginSnapshotList)
}
