// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReconcileDetailListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReconcileDetailListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReconcileDetailListLogic {
	return &ReconcileDetailListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReconcileDetailListLogic) ReconcileDetailList(req *types.ReconcileDetailQuery) (resp *types.ReconcileDetailListResp, err error) {
	return reconcileDetailList(l.ctx, l.svcCtx, req)
}
