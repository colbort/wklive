// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantProductLogic {
	return &UpdateTenantProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantProductLogic) UpdateTenantProduct(req *types.UpdateTenantProductReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
