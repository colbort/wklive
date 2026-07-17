package logic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantDomainDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantDomainDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDomainDeleteLogic {
	return &SysTenantDomainDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysTenantDomainDeleteLogic) SysTenantDomainDelete(in *system.SysTenantDomainDeleteReq) (*system.RespBase, error) {
	if base, err := systemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{Base: base}, nil
	}
	current, err := l.svcCtx.TenantDomainModel.FindOne(l.ctx, in.GetId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tenantDomainInvalidResp(l.ctx), nil
		}
		return nil, err
	}
	if current.Status == models.TenantDomainStatusActive {
		count, err := l.svcCtx.TenantDomainModel.CountActive(l.ctx, current.TenantId, current.Id)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			return tenantDomainInvalidResp(l.ctx), nil
		}
	}
	if err := l.svcCtx.TenantDomainModel.Delete(l.ctx, current.Id); err != nil {
		return nil, err
	}
	return &system.RespBase{Base: helper.OkResp()}, nil
}
