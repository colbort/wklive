// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayChannelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantPayChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayChannelsLogic {
	return &ListTenantPayChannelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantPayChannelsLogic) ListTenantPayChannels(req *types.ListTenantPayChannelsReq) (resp *types.ListTenantPayChannelsResp, err error) {
	resp, err = logicutil.Proxy[types.ListTenantPayChannelsResp](l.ctx, req, l.svcCtx.PaymentCli.ListTenantPayChannels)
	if err != nil || resp == nil || resp.Code != 200 {
		return resp, err
	}
	names := loadTenantNames(l.ctx, l.svcCtx, uniqueTenantIDs(resp.Data, func(item types.TenantPayChannel) int64 {
		return item.TenantId
	}))
	for index := range resp.Data {
		resp.Data[index].TenantName = names[resp.Data[index].TenantId]
	}
	return resp, nil
}
