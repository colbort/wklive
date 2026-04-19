package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/common"
	"wklive/proto/itick"
	"wklive/proto/system"
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
	items, _, err := l.svcCtx.ItickTenantCategoryModel.FindPage(l.ctx, in.TenantId, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	tenants, err := l.svcCtx.SystemCli.SysTenantList(l.ctx, &system.SysTenantListReq{
		Page: &common.PageReq{
			Cursor: 0,
			Limit:  100,
		},
	})
	if err != nil {
		return nil, err
	}
	tenantMap := make(map[int64]*system.SysTenantItem, len(tenants.Data))
	for _, tenant := range tenants.Data {
		tenantMap[tenant.Id] = tenant
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
	filtered := make([]*itick.ItickTenantCategory, 0)
	total := int64(0)
	for _, item := range items {
		category := categoryMap[item.CategoryId]
		tenant := tenantMap[item.TenantId]
		if category == nil {
			continue
		}
		if in.CategoryType > 0 && int64(in.CategoryType) != category.CategoryType {
			continue
		}
		if !statusMatches(in.Status, item.Enabled) {
			continue
		}
		if !statusMatches(in.VisibleStatus, item.AppVisible) {
			continue
		}

		total++
		if item.Id <= in.Page.Cursor || int64(len(filtered)) >= limit {
			continue
		}
		filtered = append(filtered, toTenantCategoryProto(item, category, tenant))
	}

	lastID := int64(0)
	if len(filtered) > 0 {
		lastID = filtered[len(filtered)-1].Id
	}

	return &itick.ListTenantCategoriesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(filtered), total, lastID),
		Data: filtered,
	}, nil
}
