package logic

import (
	"context"
	"sort"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVisibleCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVisibleCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVisibleCategoriesLogic {
	return &ListVisibleCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取允许显示的产品类型
func (l *ListVisibleCategoriesLogic) ListVisibleCategories(in *itick.ListVisibleCategoriesReq) (*itick.ListVisibleCategoriesResp, error) {
	items, err := collectTenantCategories(l.ctx, l.svcCtx.ItickTenantCategoryModel, 0)
	if err != nil {
		return nil, err
	}

	categories, err := l.svcCtx.ItickCategoryModel.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}
	categoryMap := make(map[int64]*models.TItickCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.Id] = category
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			return items[i].Id < items[j].Id
		}
		return items[i].Sort < items[j].Sort
	})

	data := make([]*itick.ItickTenantCategory, 0)
	for _, item := range items {
		category := categoryMap[item.CategoryId]
		if category == nil {
			continue
		}
		if item.Enabled != 1 || item.AppVisible != 1 {
			continue
		}
		if category.Enabled != 1 || category.AppVisible != 1 {
			continue
		}
		data = append(data, toTenantCategoryProto(item, category))
	}

	return &itick.ListVisibleCategoriesResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
