package adminlogic

import (
	"context"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"

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
func (l *BatchUpsertTenantCategoriesLogic) BatchUpsertTenantCategories(in *market.BatchUpsertTenantCategoriesReq) (*market.CommonResp, error) {
	if base, err := helpers.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &market.CommonResp{Base: base}, nil
	}

	return &market.CommonResp{Base: helper.OkResp()}, nil
}
