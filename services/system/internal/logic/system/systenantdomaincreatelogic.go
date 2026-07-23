package systemlogic

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

type SysTenantDomainCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantDomainCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDomainCreateLogic {
	return &SysTenantDomainCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysTenantDomainCreateLogic) SysTenantDomainCreate(in *system.SysTenantDomainCreateReq) (*system.RespBase, error) {
	if base, err := systemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{Base: base}, nil
	}
	origin, ok := normalizeTenantDomainOrigin(in.GetOrigin())
	if in.GetTenantId() <= 0 || !ok || !validTenantDomainStatus(in.GetStatus()) {
		return tenantDomainInvalidResp(l.ctx), nil
	}
	if _, err := l.svcCtx.TenantMode.FindOne(l.ctx, in.GetTenantId()); err != nil {
		return tenantDomainInvalidResp(l.ctx), nil
	}
	if _, err := l.svcCtx.TenantDomainModel.FindByTenantOrigin(l.ctx, in.GetTenantId(), origin); err == nil {
		return tenantDomainInvalidResp(l.ctx), nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if in.GetStatus() != system.TenantDomainStatus_TENANT_DOMAIN_STATUS_ACTIVE {
		count, err := l.svcCtx.TenantDomainModel.CountActive(l.ctx, in.GetTenantId(), 0)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			return tenantDomainInvalidResp(l.ctx), nil
		}
	}
	now := time.Now().UnixMilli()
	err := l.svcCtx.TenantDomainModel.Insert(l.ctx, &models.SysTenantDomain{TenantId: in.GetTenantId(), Origin: origin,
		Status: int64(in.GetStatus()), Priority: in.GetPriority(), CreateTimes: now, UpdateTimes: now})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{Base: helper.OkResp()}, nil
}
