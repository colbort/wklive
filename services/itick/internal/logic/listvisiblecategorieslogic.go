package logic

import (
	"context"
	"sort"

	"wklive/common/pageutil"
	"wklive/proto/itick"
	"wklive/proto/system"
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
	detail, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &in.TenantCode,
	})
	if err != nil {
		return nil, err
	}
	items, total, err := l.svcCtx.ItickTenantCategoryModel.FindPage(l.ctx, detail.Data.Id, in.Page.Cursor, in.Page.Limit)
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

	limit := pageutil.NormalizeLimit(in.Page.Limit)
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
		if item.Id <= in.Page.Cursor || int64(len(data)) >= limit {
			continue
		}
		data = append(data, toTenantCategoryProto(item, category, detail.Data))
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = data[len(data)-1].Id
	}

	return &itick.ListVisibleCategoriesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(data), total, lastID),
		Data: data,
	}, nil
}
