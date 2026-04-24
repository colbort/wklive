package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantConfigDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantConfigDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantConfigDetailLogic {
	return &TenantConfigDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取系统配置详情
func (l *TenantConfigDetailLogic) TenantConfigDetail(in *tenant.TenantConfigDetailReq) (*tenant.TenantConfigDetailResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantConfigDetailResp{}, nil
}
