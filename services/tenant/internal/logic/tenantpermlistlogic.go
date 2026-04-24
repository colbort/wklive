package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantPermListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantPermListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantPermListLogic {
	return &TenantPermListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取权限列表
func (l *TenantPermListLogic) TenantPermList(in *tenant.Empty) (*tenant.TenantPermListResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantPermListResp{}, nil
}
