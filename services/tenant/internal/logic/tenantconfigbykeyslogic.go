package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantConfigByKeysLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantConfigByKeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantConfigByKeysLogic {
	return &TenantConfigByKeysLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取系统配置根据keys
func (l *TenantConfigByKeysLogic) TenantConfigByKeys(in *tenant.TenantConfigByKeysReq) (*tenant.TenantConfigByKeysResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantConfigByKeysResp{}, nil
}
