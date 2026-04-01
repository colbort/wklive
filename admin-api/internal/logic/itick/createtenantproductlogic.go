// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantProductLogic {
	return &CreateTenantProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantProductLogic) CreateTenantProduct(req *types.CreateTenantProductReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
