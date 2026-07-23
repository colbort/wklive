package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantDomainListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantDomainListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDomainListLogic {
	return &SysTenantDomainListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysTenantDomainListLogic) SysTenantDomainList(in *system.SysTenantDomainListReq) (*system.SysTenantDomainListResp, error) {
	if in.GetTenantId() <= 0 {
		return &system.SysTenantDomainListResp{Base: tenantDomainInvalidResp(l.ctx).Base}, nil
	}
	items, err := l.svcCtx.TenantDomainModel.FindAllByTenant(l.ctx, in.GetTenantId())
	if err != nil {
		return nil, err
	}
	data := make([]*system.SysTenantDomainItem, 0, len(items))
	for _, item := range items {
		data = append(data, &system.SysTenantDomainItem{Id: item.Id, TenantId: item.TenantId, Origin: item.Origin,
			Status: system.TenantDomainStatus(item.Status), Priority: item.Priority, CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes})
	}
	return &system.SysTenantDomainListResp{Base: helper.OkResp(), Data: data}, nil
}
