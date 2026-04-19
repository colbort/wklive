package logic

import (
	"context"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/common"
	"wklive/proto/itick"
	"wklive/proto/system"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantProductsLogic {
	return &ListTenantProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户产品列表
func (l *ListTenantProductsLogic) ListTenantProducts(in *itick.ListTenantProductsReq) (*itick.ListTenantProductsResp, error) {
	items, _, err := l.svcCtx.ItickTenantProductModel.FindPage(l.ctx, in.TenantId, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	productIDs := make([]int64, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductId)
	}

	products, err := collectProductsByIDs(l.ctx, l.svcCtx.ItickProductModel, productIDs)
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
	limit := pageutil.NormalizeLimit(in.Page.Limit)
	market := strings.TrimSpace(in.Market)
	filtered := make([]*itick.ItickTenantProduct, 0)
	total := int64(0)
	for _, item := range items {
		product := products[item.ProductId]
		tenant := tenantMap[item.TenantId]
		if product == nil {
			continue
		}
		if in.CategoryType > 0 && int64(in.CategoryType) != product.CategoryType {
			continue
		}
		if market != "" && product.Market != market {
			continue
		}
		if !statusMatches(in.Status, item.Enabled) {
			continue
		}
		if !statusMatches(in.VisibleStatus, item.AppVisible) {
			continue
		}
		if !keywordMatches(in.Keyword, product.Symbol, product.Code, product.Name, product.DisplayName, product.CategoryName) {
			continue
		}

		total++
		if item.Id <= in.Page.Cursor || int64(len(filtered)) >= limit {
			continue
		}
		filtered = append(filtered, toTenantProductProto(item, product, tenant))
	}

	lastID := int64(0)
	if len(filtered) > 0 {
		lastID = filtered[len(filtered)-1].Id
	}

	return &itick.ListTenantProductsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(filtered), total, lastID),
		Data: filtered,
	}, nil
}
