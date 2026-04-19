package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchUpsertTenantCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchUpsertTenantCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpsertTenantCategoriesLogic {
	return &BatchUpsertTenantCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量更新租户产品类型，已关联的修改状态、排序和备注，未关联的新增，未提交的删除
func (l *BatchUpsertTenantCategoriesLogic) BatchUpsertTenantCategories(in *itick.BatchUpsertTenantCategoriesReq) (*itick.AdminCommonResp, error) {

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
