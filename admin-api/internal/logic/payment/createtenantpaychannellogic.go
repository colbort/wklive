// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantPayChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantPayChannelLogic {
	return &CreateTenantPayChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantPayChannelLogic) CreateTenantPayChannel(req *types.CreateTenantPayChannelReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
