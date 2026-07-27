package adminlogic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantCategoriesLogic {
	return &ListTenantCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户产品类型列表
func (l *ListTenantCategoriesLogic) ListTenantCategories(in *itick.ListTenantCategoriesReq) (*itick.ListTenantCategoriesResp, error) {
	items, total, err := l.svcCtx.ItickTenantCategoryModel.FindPage(l.ctx, in.TenantId, in.Page.Cursor, in.Page.Limit)
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

	limit := pageutil.NormalizeLimit(in.Page.Limit)
	data := make([]*itick.ItickTenantCategory, 0)
	for _, item := range items {
		category := categoryMap[item.CategoryId]
		if category == nil {
			continue
		}
		if in.CategoryType > 0 && int64(in.CategoryType) != category.CategoryType {
			continue
		}
		if !statusMatches(int32(in.Enabled), item.Enabled) {
			continue
		}
		if !statusMatches(int32(in.VisibleStatus), item.AppVisible) {
			continue
		}

		if item.Id <= in.Page.Cursor || int64(len(data)) >= limit {
			continue
		}
		data = append(data, toTenantCategoryProto(item, category))
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = data[len(data)-1].Id
	}

	return &itick.ListTenantCategoriesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(data), total, lastID),
		Data: data,
	}, nil
}
