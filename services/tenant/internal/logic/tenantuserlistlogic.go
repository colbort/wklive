package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantUserListLogic {
	return &TenantUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户
func (l *TenantUserListLogic) TenantUserList(in *tenant.TenantUserListReq) (*tenant.TenantUserListResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantUserListResp{}, nil
}
