package internallogic

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

type ResolveTenantDomainLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveTenantDomainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveTenantDomainLogic {
	return &ResolveTenantDomainLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据租户和来源域名解析游客迁移目标
func (l *ResolveTenantDomainLogic) ResolveTenantDomain(in *system.ResolveTenantDomainReq) (*system.ResolveTenantDomainResp, error) {
	resp := &system.ResolveTenantDomainResp{Base: helper.OkResp()}
	if in.GetTenantId() <= 0 || in.GetSourceOrigin() == "" {
		return resp, nil
	}

	source, err := l.svcCtx.TenantDomainModel.FindByTenantOrigin(l.ctx, in.GetTenantId(), in.GetSourceOrigin())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return resp, nil
		}
		return nil, err
	}
	resp.SourceStatus = system.TenantDomainStatus(source.Status)
	if source.Status != models.TenantDomainStatusRetired {
		return resp, nil
	}

	target, err := l.svcCtx.TenantDomainModel.FindHighestPriorityActive(l.ctx, in.GetTenantId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return resp, nil
		}
		return nil, err
	}
	resp.TargetOrigin = target.Origin
	return resp, nil
}
