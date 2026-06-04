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

type ListTenantPayChannelRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantPayChannelRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayChannelRulesLogic {
	return &ListTenantPayChannelRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantPayChannelRulesLogic) ListTenantPayChannelRules(req *types.ListTenantPayChannelRulesReq) (resp *types.ListTenantPayChannelRulesResp, err error) {
	return logicutil.Proxy[types.ListTenantPayChannelRulesResp](l.ctx, req, l.svcCtx.PaymentCli.ListTenantPayChannelRules)
}
