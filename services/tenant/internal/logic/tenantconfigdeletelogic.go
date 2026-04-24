package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantConfigDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantConfigDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantConfigDeleteLogic {
	return &TenantConfigDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除系统配置
func (l *TenantConfigDeleteLogic) TenantConfigDelete(in *tenant.TenantConfigDeleteReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
