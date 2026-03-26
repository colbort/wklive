// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantPayAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayAccountLogic {
	return &UpdateTenantPayAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantPayAccountLogic) UpdateTenantPayAccount(req *types.UpdateTenantPayAccountReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
