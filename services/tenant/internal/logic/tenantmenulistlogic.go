package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantMenuListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTenantMenuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantMenuListLogic {
	return &TenantMenuListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取菜单列表
func (l *TenantMenuListLogic) TenantMenuList(in *tenant.TenantMenuListReq) (*tenant.TenantMenuListResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.TenantMenuListResp{}, nil
}
