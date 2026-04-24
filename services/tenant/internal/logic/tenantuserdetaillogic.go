package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantUserDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantUserDetailLogic {
	return &TenantUserDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户详情
func (l *TenantUserDetailLogic) TenantUserDetail(in *tenant.TenantUserDetailReq) (*tenant.TenantUserDetailResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantUserDetailResp{}, nil
}
