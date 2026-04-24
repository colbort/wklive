package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantConfigListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantConfigListLogic {
	return &TenantConfigListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取系统配置列表
func (l *TenantConfigListLogic) TenantConfigList(in *tenant.TenantConfigListReq) (*tenant.TenantConfigListResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantConfigListResp{}, nil
}
