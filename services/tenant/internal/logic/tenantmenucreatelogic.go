package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantMenuCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantMenuCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantMenuCreateLogic {
	return &TenantMenuCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 菜单
func (l *TenantMenuCreateLogic) TenantMenuCreate(in *tenant.TenantMenuCreateReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
