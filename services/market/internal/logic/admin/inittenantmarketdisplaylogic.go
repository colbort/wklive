package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/market"
	"wklive/services/market/internal/logic/helpers"
	"wklive/services/market/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitTenantMarketDisplayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitTenantMarketDisplayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantMarketDisplayLogic {
	return &InitTenantMarketDisplayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化租户展示配置
func (l *InitTenantMarketDisplayLogic) InitTenantMarketDisplay(in *market.InitTenantMarketDisplayReq) (*market.InitTenantMarketDisplayResp, error) {
	if base, err := helpers.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &market.InitTenantMarketDisplayResp{Base: base}, nil
	}

	// now := cutils.NowMillis()
	categoryCount := int64(0)
	productCount := int64(0)

	// categories, err := l.svcCtx.MarketCategoryModel.FindAll(l.ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, category := range categories {
	// 	exist, err := l.svcCtx.MarketTenantCategoryModel.FindOneByTenantIdCategoryId(l.ctx, in.TenantId, category.Id)
	// 	if err != nil && !errors.Is(err, models.ErrNotFound) {
	// 		return nil, err
	// 	}

	// 	if exist == nil {
	// 		_, err = l.svcCtx.MarketTenantCategoryModel.Insert(l.ctx, &models.TMarketTenantCategory{
	// 			TenantId:    in.TenantId,
	// 			CategoryId:  category.Id,
	// 			Enabled:     category.Enabled,
	// 			AppVisible:  category.AppVisible,
	// 			Sort:        category.Sort,
	// 			Remark:      category.Remark,
	// 			CreateTimes: now,
	// 			UpdateTimes: now,
	// 		})
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		categoryCount++
	// 		continue
	// 	}

	// 	if in.Overwrite == 1 {
	// 		exist.Enabled = category.Enabled
	// 		exist.AppVisible = category.AppVisible
	// 		exist.Sort = category.Sort
	// 		exist.Remark = category.Remark
	// 		exist.UpdateTimes = now
	// 		if err := l.svcCtx.MarketTenantCategoryModel.Update(l.ctx, exist); err != nil {
	// 			return nil, err
	// 		}
	// 		categoryCount++
	// 	}
	// }

	// products, err := collectProducts(l.ctx, l.svcCtx.MarketProductModel)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, product := range products {
	// 	exist, err := l.svcCtx.MarketTenantProductModel.FindOneByTenantIdProductId(l.ctx, in.TenantId, product.Id)
	// 	if err != nil && !errors.Is(err, models.ErrNotFound) {
	// 		return nil, err
	// 	}

	// 	if exist == nil {
	// 		_, err = l.svcCtx.MarketTenantProductModel.Insert(l.ctx, &models.TMarketTenantProduct{
	// 			TenantId:    in.TenantId,
	// 			ProductId:   product.Id,
	// 			Enabled:     product.Enabled,
	// 			AppVisible:  product.AppVisible,
	// 			Sort:        product.Sort,
	// 			Remark:      product.Remark,
	// 			CreateTimes: now,
	// 			UpdateTimes: now,
	// 		})
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		productCount++
	// 		continue
	// 	}

	// 	if in.Overwrite == 1 {
	// 		exist.Enabled = product.Enabled
	// 		exist.AppVisible = product.AppVisible
	// 		exist.Sort = product.Sort
	// 		exist.Remark = product.Remark
	// 		exist.UpdateTimes = now
	// 		if err := l.svcCtx.MarketTenantProductModel.Update(l.ctx, exist); err != nil {
	// 			return nil, err
	// 		}
	// 		productCount++
	// 	}
	// }

	return &market.InitTenantMarketDisplayResp{
		Base: helper.OkResp(),
		Data: &market.InitTenantMarketDisplayData{
			CategoryCount: categoryCount,
			ProductCount:  productCount,
		},
	}, nil
}
