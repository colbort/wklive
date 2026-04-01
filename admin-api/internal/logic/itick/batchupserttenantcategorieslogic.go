// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchUpsertTenantCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchUpsertTenantCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpsertTenantCategoriesLogic {
	return &BatchUpsertTenantCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchUpsertTenantCategoriesLogic) BatchUpsertTenantCategories(req *types.BatchUpsertTenantCategoriesReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
