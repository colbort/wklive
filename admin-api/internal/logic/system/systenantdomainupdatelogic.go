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

type SysTenantDomainUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysTenantDomainUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDomainUpdateLogic {
	return &SysTenantDomainUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysTenantDomainUpdateLogic) SysTenantDomainUpdate(req *types.SysTenantDomainUpdateReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.SystemCli.SysTenantDomainUpdate)
}
