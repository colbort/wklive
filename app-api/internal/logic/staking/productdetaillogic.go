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

type ProductDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductDetailLogic {
	return &ProductDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductDetailLogic) ProductDetail(req *types.ProductDetailReq) (resp *types.ProductDetailResp, err error) {
	ctx, err := resolveTenantContext(l.ctx, l.svcCtx)
	if err != nil {
		return logicutil.SystemErrorResp[types.ProductDetailResp](l.ctx, err)
	}
	return logicutil.Proxy[types.ProductDetailResp](ctx, req, l.svcCtx.StakingCli.ProductDetail)
}
