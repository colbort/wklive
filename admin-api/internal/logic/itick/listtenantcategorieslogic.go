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

type ListTenantCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantCategoriesLogic {
	return &ListTenantCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantCategoriesLogic) ListTenantCategories(req *types.ListTenantCategoriesReq) (resp *types.ListTenantCategoriesResp, err error) {
	return logicutil.Proxy[types.ListTenantCategoriesResp](l.ctx, req, l.svcCtx.ItickCli.ListTenantCategories)
}
