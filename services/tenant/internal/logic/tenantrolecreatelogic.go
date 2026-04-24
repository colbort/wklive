package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantRoleCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantRoleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantRoleCreateLogic {
	return &TenantRoleCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建角色
func (l *TenantRoleCreateLogic) TenantRoleCreate(in *tenant.TenantRoleCreateReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
