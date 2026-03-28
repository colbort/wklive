package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchUpsertTenantProductCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchUpsertTenantProductCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpsertTenantProductCategoriesLogic {
	return &BatchUpsertTenantProductCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量更新租户产品类型，已关联的修改状态、排序和备注，未关联的新增，未提交的删除
func (l *BatchUpsertTenantProductCategoriesLogic) BatchUpsertTenantProductCategories(in *itick.BatchUpsertTenantProductCategoriesReq) (*itick.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &itick.AdminCommonResp{}, nil
}
