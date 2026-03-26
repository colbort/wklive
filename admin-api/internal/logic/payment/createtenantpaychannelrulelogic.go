// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantPayChannelRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantPayChannelRuleLogic {
	return &CreateTenantPayChannelRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantPayChannelRuleLogic) CreateTenantPayChannelRule(req *types.CreateTenantPayChannelRuleReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
