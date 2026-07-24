// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	pb "wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReconcileBatchListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReconcileBatchListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReconcileBatchListLogic {
	return &ReconcileBatchListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReconcileBatchListLogic) ReconcileBatchList(req *types.PageQuery) (resp *types.ReconcileListResp, err error) {
	tenantID, err := logicutil.TenantID(l.ctx)
	if err != nil {
		return nil, err
	}
	out, err := l.svcCtx.LiquidityCli.GetReconcileBatchList(l.ctx, &pb.GetReconcileBatchListReq{
		TenantId: tenantID, Status: pb.ReconcileStatus(req.Status), Cursor: req.Cursor, Limit: listLimit(req.Limit),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.ReconcileListResp](out), nil
}
