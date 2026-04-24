package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantConfigUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantConfigUpdateLogic {
	return &TenantConfigUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新系统配置
func (l *TenantConfigUpdateLogic) TenantConfigUpdate(in *tenant.TenantConfigUpdateReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
