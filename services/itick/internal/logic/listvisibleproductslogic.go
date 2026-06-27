package logic

import (
	"context"
	"sort"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/itick"
	"wklive/proto/system"
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
	tenantCode, err := utils.GetTenantCodeFromMd(l.ctx)
	if err != nil || tenantCode == "" {
		return &itick.ListVisibleProductsResp{
			Base: helper.ErrResp(i18n.InvalidRequest, i18n.Translate(i18n.InvalidRequest, l.ctx)),
		}, nil
	}
	detail, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &tenantCode,
	})
	if err != nil {
		return nil, err
	}
	items, total, err := l.svcCtx.ItickTenantProductModel.FindPage(l.ctx, models.TenantProductPageFilter{
		TenantId:     detail.Data.Id,
		CategoryType: int64(in.CategoryType),
		Enabled:      1,
		AppVisible:   1,
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

	products, err := collectProductsByIDs(l.ctx, l.svcCtx.ItickProductModel, productIDs)
	if err != nil {
		return nil, err
	}

	data := make([]*itick.ItickTenantProduct, 0)
	for _, item := range items {
		product := products[item.ProductId]
		if product == nil {
			continue
		}
		data = append(data, toTenantProductProto(item, product, detail.Data))
	}

	return &itick.ListVisibleProductsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(data), total, lastID),
		Data: data,
	}, nil
}
