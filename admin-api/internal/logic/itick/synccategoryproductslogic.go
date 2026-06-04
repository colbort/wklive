// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncCategoryProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncCategoryProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncCategoryProductsLogic {
	return &SyncCategoryProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncCategoryProductsLogic) SyncCategoryProducts(req *types.SyncCategoryProductsReq) (resp *types.SyncCategoryProductsResp, err error) {
	return logicutil.Proxy[types.SyncCategoryProductsResp](l.ctx, req, l.svcCtx.ItickCli.SyncCategoryProducts)
}
