// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package market

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
	resp, err = logicutil.Proxy[types.ListTenantProductsResp](l.ctx, req, l.svcCtx.MarketCli.ListTenantProducts)
	if err != nil || resp == nil || resp.Code != 200 {
		return resp, err
	}
	ids := make([]int64, 0, len(resp.Data))
	for _, item := range resp.Data {
		ids = append(ids, item.TenantId)
	}
	names, err := loadTenantNames(l.ctx, l.svcCtx, ids)
	if err != nil {
		return nil, err
	}
	for i := range resp.Data {
		resp.Data[i].TenantName = names[resp.Data[i].TenantId]
	}
	return resp, nil
}
