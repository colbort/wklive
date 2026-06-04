// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantPayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayPlatformLogic {
	return &GetTenantPayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantPayPlatformLogic) GetTenantPayPlatform(req *types.GetTenantPayPlatformReq) (resp *types.GetTenantPayPlatformResp, err error) {
	return logicutil.Proxy[types.GetTenantPayPlatformResp](l.ctx, req, l.svcCtx.PaymentCli.GetTenantPayPlatform)
}
