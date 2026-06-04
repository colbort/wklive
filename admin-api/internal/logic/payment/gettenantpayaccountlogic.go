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

type GetTenantPayAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayAccountLogic {
	return &GetTenantPayAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantPayAccountLogic) GetTenantPayAccount(req *types.GetTenantPayAccountReq) (resp *types.GetTenantPayAccountResp, err error) {
	return logicutil.Proxy[types.GetTenantPayAccountResp](l.ctx, req, l.svcCtx.PaymentCli.GetTenantPayAccount)
}
