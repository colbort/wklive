// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitTenantItickDisplayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitTenantItickDisplayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantItickDisplayLogic {
	return &InitTenantItickDisplayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitTenantItickDisplayLogic) InitTenantItickDisplay(req *types.InitTenantItickDisplayReq) (resp *types.InitTenantItickDisplayResp, err error) {
	return logicutil.Proxy[types.InitTenantItickDisplayResp](l.ctx, req, l.svcCtx.ItickCli.InitTenantItickDisplay)
}
