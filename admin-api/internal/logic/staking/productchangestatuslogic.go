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

type ProductChangeStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProductChangeStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductChangeStatusLogic {
	return &ProductChangeStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductChangeStatusLogic) ProductChangeStatus(req *types.AdminProductChangeStatusReq) (resp *types.AdminProductChangeStatusResp, err error) {
	return logicutil.Proxy[types.AdminProductChangeStatusResp](l.ctx, req, l.svcCtx.StakingCli.ProductChangeStatus)
}
