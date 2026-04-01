// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantCategoryLogic {
	return &CreateTenantCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantCategoryLogic) CreateTenantCategory(req *types.CreateTenantCategoryReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
