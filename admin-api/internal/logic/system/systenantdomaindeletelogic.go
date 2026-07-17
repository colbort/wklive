// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantDomainDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysTenantDomainDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDomainDeleteLogic {
	return &SysTenantDomainDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysTenantDomainDeleteLogic) SysTenantDomainDelete(req *types.SysTenantDomainDeleteReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.SystemCli.SysTenantDomainDelete)
}
