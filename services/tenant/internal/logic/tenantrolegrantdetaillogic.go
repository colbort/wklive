package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantRoleGrantDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantRoleGrantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantRoleGrantDetailLogic {
	return &TenantRoleGrantDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取角色授权详情
func (l *TenantRoleGrantDetailLogic) TenantRoleGrantDetail(in *tenant.TenantRoleGrantDetailReq) (*tenant.TenantRoleGrantDetailResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantRoleGrantDetailResp{}, nil
}
