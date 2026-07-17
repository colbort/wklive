package system

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantDomainGuestMigrationStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysTenantDomainGuestMigrationStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDomainGuestMigrationStatsLogic {
	return &SysTenantDomainGuestMigrationStatsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *SysTenantDomainGuestMigrationStatsLogic) SysTenantDomainGuestMigrationStats(req *types.SysTenantDomainGuestMigrationStatsReq) (*types.SysTenantDomainGuestMigrationStatsResp, error) {
	return logicutil.Proxy[types.SysTenantDomainGuestMigrationStatsResp](l.ctx, &user.GuestDomainMigrationStatsReq{
		TenantId: req.TenantId, SourceOrigin: req.SourceOrigin,
	}, l.svcCtx.UserCli.GuestDomainMigrationStats)
}
