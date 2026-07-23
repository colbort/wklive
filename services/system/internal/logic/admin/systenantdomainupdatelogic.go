package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantDomainUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantDomainUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDomainUpdateLogic {
	return &SysTenantDomainUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysTenantDomainUpdateLogic) SysTenantDomainUpdate(in *system.SysTenantDomainUpdateReq) (*system.RespBase, error) {
	if base, err := systemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{Base: base}, nil
	}
	origin, ok := normalizeTenantDomainOrigin(in.GetOrigin())
	if in.GetId() <= 0 || !ok || !validTenantDomainStatus(in.GetStatus()) {
		return tenantDomainInvalidResp(l.ctx), nil
	}
	current, err := l.svcCtx.TenantDomainModel.FindOne(l.ctx, in.GetId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tenantDomainInvalidResp(l.ctx), nil
		}
		return nil, err
	}
	if duplicate, err := l.svcCtx.TenantDomainModel.FindByTenantOrigin(l.ctx, current.TenantId, origin); err == nil && duplicate.Id != current.Id {
		return tenantDomainInvalidResp(l.ctx), nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if current.Status == models.TenantDomainStatusActive && in.GetStatus() != system.TenantDomainStatus_TENANT_DOMAIN_STATUS_ACTIVE {
		count, err := l.svcCtx.TenantDomainModel.CountActive(l.ctx, current.TenantId, current.Id)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			return tenantDomainInvalidResp(l.ctx), nil
		}
	}
	current.Origin, current.Status, current.Priority, current.UpdateTimes = origin, int64(in.GetStatus()), in.GetPriority(), time.Now().UnixMilli()
	if err := l.svcCtx.TenantDomainModel.Update(l.ctx, current); err != nil {
		return nil, err
	}
	return &system.RespBase{Base: helper.OkResp()}, nil
}
