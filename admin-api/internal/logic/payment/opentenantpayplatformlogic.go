// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpenTenantPayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOpenTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenTenantPayPlatformLogic {
	return &OpenTenantPayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OpenTenantPayPlatformLogic) OpenTenantPayPlatform(req *types.OpenTenantPayPlatformReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
