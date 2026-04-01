// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchUpsertTenantProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchUpsertTenantProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpsertTenantProductsLogic {
	return &BatchUpsertTenantProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchUpsertTenantProductsLogic) BatchUpsertTenantProducts(req *types.BatchUpsertTenantProductsReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
