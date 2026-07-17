package logic

import (
	"context"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GuestDomainMigrationStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGuestDomainMigrationStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GuestDomainMigrationStatsLogic {
	return &GuestDomainMigrationStatsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GuestDomainMigrationStatsLogic) GuestDomainMigrationStats(in *user.GuestDomainMigrationStatsReq) (*user.GuestDomainMigrationStatsResp, error) {
	origin, ok := normalizeTransferOrigin(strings.TrimSpace(in.GetSourceOrigin()))
	if in.GetTenantId() <= 0 || !ok {
		return &user.GuestDomainMigrationStatsResp{Base: helper.ErrResp(i18n.InvalidRequest, i18n.Translate(i18n.InvalidRequest, l.ctx))}, nil
	}
	now := time.Now()
	stats, err := l.svcCtx.UserModel.GetGuestDomainMigrationStats(
		l.ctx, in.GetTenantId(), origin,
		now.AddDate(0, 0, -7).UnixMilli(),
		now.AddDate(0, 0, -30).UnixMilli(),
		now.AddDate(0, 0, -90).UnixMilli(),
	)
	if err != nil {
		return nil, err
	}
	return &user.GuestDomainMigrationStatsResp{
		Base: helper.OkResp(),
		Data: &user.GuestDomainMigrationStatsData{
			NotMigratedCount:          stats.NotMigratedCount,
			ActiveWithinWeekCount:     stats.Active7dCount,
			ActiveWeekToMonthCount:    stats.Active8To30dCount,
			ActiveMonthToQuarterCount: stats.Active31To90dCount,
			InactiveOverQuarterCount:  stats.InactiveOver90dCount,
		},
	}, nil
}
