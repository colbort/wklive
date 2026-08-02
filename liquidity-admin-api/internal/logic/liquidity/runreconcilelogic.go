// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunReconcileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRunReconcileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunReconcileLogic {
	return &RunReconcileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RunReconcileLogic) RunReconcile(req *types.RunReconcileReq) (resp *types.ReconcileResp, err error) {
	return runReconcile(l.ctx, l.svcCtx, req)
}
