// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTenantPayChannelRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantPayChannelRuleLogic {
	return &DeleteTenantPayChannelRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTenantPayChannelRuleLogic) DeleteTenantPayChannelRule() (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
