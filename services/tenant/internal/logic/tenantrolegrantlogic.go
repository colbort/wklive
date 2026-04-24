package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantRoleGrantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantRoleGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantRoleGrantLogic {
	return &TenantRoleGrantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 角色授权
func (l *TenantRoleGrantLogic) TenantRoleGrant(in *tenant.TenantRoleGrantReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
