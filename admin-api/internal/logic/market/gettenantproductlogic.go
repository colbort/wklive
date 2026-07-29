// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package market

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantProductLogic {
	return &GetTenantProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantProductLogic) GetTenantProduct(req *types.GetTenantProductReq) (resp *types.GetTenantProductResp, err error) {
	resp, err = logicutil.Proxy[types.GetTenantProductResp](l.ctx, req, l.svcCtx.MarketCli.GetTenantProduct)
	if err != nil || resp == nil || resp.Code != 200 {
		return resp, err
	}
	names, err := loadTenantNames(l.ctx, l.svcCtx, []int64{resp.Data.TenantId})
	if err != nil {
		return nil, err
	}
	resp.Data.TenantName = names[resp.Data.TenantId]
	return resp, nil
}
