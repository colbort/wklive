// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantPayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayPlatformLogic {
	return &UpdateTenantPayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantPayPlatformLogic) UpdateTenantPayPlatform(req *types.UpdateTenantPayPlatformReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
