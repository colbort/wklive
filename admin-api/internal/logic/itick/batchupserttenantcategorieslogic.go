// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/itick"

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
	data := make([]*itick.TenantCategoryItem, 0, len(req.Data))
	for _, item := range req.Data {
		data = append(data, &itick.TenantCategoryItem{
			Id:         item.Id,
			CategoryId: item.CategoryId,
			Enabled:    item.Enabled,
			AppVisible: item.AppVisible,
			Sort:       item.Sort,
			Remark:     item.Remark,
		})
	}

	result, err := l.svcCtx.ItickCli.BatchUpsertTenantCategories(l.ctx, &itick.BatchUpsertTenantCategoriesReq{
		TenantId: req.TenantId,
		Data:     data,
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}
