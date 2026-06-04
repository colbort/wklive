// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantProductsLogic {
	return &ListTenantProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantProductsLogic) ListTenantProducts(req *types.ListTenantProductsReq) (resp *types.ListTenantProductsResp, err error) {
	return logicutil.Proxy[types.ListTenantProductsResp](l.ctx, req, l.svcCtx.ItickCli.ListTenantProducts)
}
