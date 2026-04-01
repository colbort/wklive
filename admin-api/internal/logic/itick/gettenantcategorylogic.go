// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantCategoryLogic {
	return &GetTenantCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantCategoryLogic) GetTenantCategory(req *types.GetTenantCategoryReq) (resp *types.GetTenantCategoryResp, err error) {
	// todo: add your logic here and delete this line

	return
}
