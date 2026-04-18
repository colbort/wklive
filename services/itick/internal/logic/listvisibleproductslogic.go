package logic

import (
	"context"
	"sort"
	"strings"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVisibleProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVisibleProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVisibleProductsLogic {
	return &ListVisibleProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取允许显示的产品
func (l *ListVisibleProductsLogic) ListVisibleProducts(in *itick.ListVisibleProductsReq) (*itick.ListVisibleProductsResp, error) {
	items, err := collectTenantProducts(l.ctx, l.svcCtx.ItickTenantProductModel, 0)
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

	sort.Slice(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			return items[i].Id < items[j].Id
		}
		return items[i].Sort < items[j].Sort
	})

	market := strings.TrimSpace(in.Market)
	data := make([]*itick.ItickTenantProduct, 0)
	for _, item := range items {
		product := productMap[item.ProductId]
		if product == nil {
			continue
		}
		if item.Enabled != 1 || item.AppVisible != 1 {
			continue
		}
		if product.Enabled != 1 || product.AppVisible != 1 {
			continue
		}
		if in.CategoryType > 0 && int64(in.CategoryType) != product.CategoryType {
			continue
		}
		if market != "" && product.Market != market {
			continue
		}
		if !keywordMatches(in.Keyword, product.Symbol, product.Code, product.Name, product.DisplayName, product.CategoryName) {
			continue
		}
		data = append(data, toTenantProductProto(item, product))
	}

	return &itick.ListVisibleProductsResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
