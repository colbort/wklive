package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantLoginLogic {
	return &TenantLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// P0
func (l *TenantLoginLogic) TenantLogin(in *tenant.TenantLoginReq) (*tenant.TenantLoginResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantLoginResp{}, nil
}
