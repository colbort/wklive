package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantConfigCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantConfigCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantConfigCreateLogic {
	return &TenantConfigCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 新增系统配置
func (l *TenantConfigCreateLogic) TenantConfigCreate(in *tenant.TenantConfigCreateReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
