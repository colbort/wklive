package applogic

import (
	"context"
	"sort"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

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
func (l *ListVisibleProductsLogic) ListVisibleProducts(in *market.ListVisibleProductsReq) (*market.ListVisibleProductsResp, error) {
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil || tenantID <= 0 {
		return &market.ListVisibleProductsResp{
			Base: helper.ErrResp(i18n.InvalidRequest, i18n.Translate(i18n.InvalidRequest, l.ctx)),
		}, nil
	}
	items, total, err := l.svcCtx.MarketTenantProductModel.FindPage(l.ctx, models.TenantProductPageFilter{
		TenantId:          tenantID,
		CategoryType:      int64(in.CategoryType),
		Enabled:           1,
		AppVisible:        1,
		ProductEnabled:    1,
		ProductAppVisible: 1,
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
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

	products, err := collectProductsByIDs(l.ctx, l.svcCtx.MarketProductModel, productIDs)
	if err != nil {
		return nil, err
	}

	data := make([]*market.MarketTenantProduct, 0)
	for _, item := range items {
		product := products[item.ProductId]
		if product == nil {
			continue
		}
		data = append(data, helpers.ToTenantProductProto(item, product))
	}

	return &market.ListVisibleProductsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(data), total, lastID),
		Data: data,
	}, nil
}
