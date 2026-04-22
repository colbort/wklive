package logic

import (
	"context"
	"sort"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/itick"
	"wklive/proto/system"
	"wklive/services/itick/internal/svc"

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
	detail, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &in.TenantCode,
	})
	if err != nil {
		return nil, err
	}
	items, total, err := l.svcCtx.ItickTenantProductModel.FindPage(l.ctx, detail.Data.Id, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			return items[i].Id < items[j].Id
		}
		return items[i].Sort < items[j].Sort
	})

	productIDs := make([]int64, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductId)
	}

	products, err := collectProductsByIDs(l.ctx, l.svcCtx.ItickProductModel, productIDs)
	if err != nil {
		return nil, err
	}

	market := strings.TrimSpace(in.Market)
	data := make([]*itick.ItickTenantProduct, 0)
	for _, item := range items {
		product := products[item.ProductId]
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
		data = append(data, toTenantProductProto(item, product, detail.Data))
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = data[len(data)-1].Id
	}

	return &itick.ListVisibleProductsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(data), total, lastID),
		Data: data,
	}, nil
}
