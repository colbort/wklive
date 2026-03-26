// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

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
	// todo: add your logic here and delete this line

	return
}
