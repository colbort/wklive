// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTenantPayChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantPayChannelLogic {
	return &DeleteTenantPayChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTenantPayChannelLogic) DeleteTenantPayChannel() (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
