// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductListLogic {
	return &ProductListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductListLogic) ProductList(req *types.ProductListReq) (resp *types.ProductListResp, err error) {
	ctx, err := resolveTenantContext(l.ctx, l.svcCtx)
	if err != nil {
		return logicutil.SystemErrorResp[types.ProductListResp](l.ctx, err)
	}
	return logicutil.Proxy[types.ProductListResp](ctx, req, l.svcCtx.StakingCli.ProductList)
}
