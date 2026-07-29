package adminlogic

import (
	"context"
	"strings"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/pageutil"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

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
func (l *ListTenantProductsLogic) ListTenantProducts(in *market.ListTenantProductsReq) (*market.ListTenantProductsResp, error) {
	items, _, err := l.svcCtx.MarketTenantProductModel.FindPage(l.ctx, models.TenantProductPageFilter{
		TenantId:     in.TenantId,
		CategoryType: int64(in.CategoryType),
		AppVisible:   int64(in.AppVisible),
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	productIDs := make([]int64, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductId)
	}

	products, err := collectProductsByIDs(l.ctx, l.svcCtx.MarketProductModel, productIDs)
	if err != nil {
		return nil, err
	}
	limit := pageutil.NormalizeLimit(in.Page.Limit)
	mt := strings.TrimSpace(in.Market)
	filtered := make([]*market.MarketTenantProduct, 0)
	total := int64(0)
	for _, item := range items {
		product := products[item.ProductId]
		if product == nil {
			continue
		}
		if in.CategoryType > 0 && int64(in.CategoryType) != product.CategoryType {
			continue
		}
		if mt != "" && product.Market != mt {
			continue
		}
		if !helpers.StatusMatches(int32(in.Enabled), item.Enabled) {
			continue
		}
		if !helpers.StatusMatches(int32(in.AppVisible), item.AppVisible) {
			continue
		}
		if !helpers.KeywordMatches(in.Keyword, product.Symbol, product.Code, product.Name, item.DisplayName, product.DisplayName, product.CategoryName) {
			continue
		}

		total++
		if item.Id <= in.Page.Cursor || int64(len(filtered)) >= limit {
			continue
		}
		filtered = append(filtered, helpers.ToTenantProductProto(item, product))
	}

	lastID := int64(0)
	if len(filtered) > 0 {
		lastID = filtered[len(filtered)-1].Id
	}

	return &market.ListTenantProductsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(filtered), total, lastID),
		Data: filtered,
	}, nil
}
