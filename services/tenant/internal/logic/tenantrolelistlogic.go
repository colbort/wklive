package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantRoleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantRoleListLogic {
	return &TenantRoleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 角色
func (l *TenantRoleListLogic) TenantRoleList(in *tenant.TenantRoleListReq) (*tenant.TenantRoleListResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantRoleListResp{}, nil
}
