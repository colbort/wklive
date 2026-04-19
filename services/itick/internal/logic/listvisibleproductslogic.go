package logic

import (
	"context"
	"sort"
	"strings"

	"wklive/common/helper"
	"wklive/proto/common"
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
	items, _, err := l.svcCtx.ItickTenantProductModel.FindPage(l.ctx, in.TenantId, in.Page.Cursor, in.Page.Limit)
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

	market := strings.TrimSpace(in.Market)
	data := make([]*itick.ItickTenantProduct, 0)
	for _, item := range items {
		product := products[item.ProductId]
		tenant := tenantMap[item.TenantId]
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
		data = append(data, toTenantProductProto(item, product, tenant))
	}

	return &itick.ListVisibleProductsResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
