package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchUpsertTenantProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchUpsertTenantProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpsertTenantProductsLogic {
	return &BatchUpsertTenantProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量更新租户产品，已关联的修改状态、排序和备注，未关联的新增，未提交的删除
func (l *BatchUpsertTenantProductsLogic) BatchUpsertTenantProducts(in *itick.BatchUpsertTenantProductsReq) (*itick.AdminCommonResp, error) {

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
