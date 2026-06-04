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

type GetTenantPayChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayChannelLogic {
	return &GetTenantPayChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantPayChannelLogic) GetTenantPayChannel(req *types.GetTenantPayChannelReq) (resp *types.GetTenantPayChannelResp, err error) {
	return logicutil.Proxy[types.GetTenantPayChannelResp](l.ctx, req, l.svcCtx.PaymentCli.GetTenantPayChannel)
}
