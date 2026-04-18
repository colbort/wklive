package logic

import (
	"context"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

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
	items, err := collectTenantProducts(l.ctx, l.svcCtx.ItickTenantProductModel, in.TenantId)
	if err != nil {
		return nil, err
	}

	products, err := collectProducts(l.ctx, l.svcCtx.ItickProductModel)
	if err != nil {
		return nil, err
	}
	productMap := make(map[int64]*models.TItickProduct, len(products))
	for _, product := range products {
		productMap[product.Id] = product
	}

	limit := pageutil.NormalizeLimit(in.Page.Limit)
	market := strings.TrimSpace(in.Market)
	filtered := make([]*itick.ItickTenantProduct, 0)
	total := int64(0)
	for _, item := range items {
		product := productMap[item.ProductId]
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
		filtered = append(filtered, toTenantProductProto(item, product))
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
