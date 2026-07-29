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

type InitTenantMarketDisplayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitTenantMarketDisplayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantMarketDisplayLogic {
	return &InitTenantMarketDisplayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitTenantMarketDisplayLogic) InitTenantMarketDisplay(req *types.InitTenantMarketDisplayReq) (resp *types.InitTenantMarketDisplayResp, err error) {
	return logicutil.Proxy[types.InitTenantMarketDisplayResp](l.ctx, req, l.svcCtx.MarketCli.InitTenantMarketDisplay)
}
