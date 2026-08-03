// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReconciliationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReconciliationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReconciliationListLogic {
	return &ReconciliationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReconciliationListLogic) ReconciliationList(req *types.ReconciliationListReq) (resp *types.ReconciliationListResp, err error) {
	return logicutil.Proxy[types.ReconciliationListResp](l.ctx, req, l.svcCtx.StakingCli.ReconciliationList)
}
