package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitTenantItickDisplayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitTenantItickDisplayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantItickDisplayLogic {
	return &InitTenantItickDisplayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化租户展示配置
func (l *InitTenantItickDisplayLogic) InitTenantItickDisplay(in *itick.InitTenantItickDisplayReq) (*itick.InitTenantItickDisplayResp, error) {
	// now := cutils.NowMillis()
	categoryCount := int64(0)
	productCount := int64(0)

	// categories, err := l.svcCtx.ItickCategoryModel.FindAll(l.ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, category := range categories {
	// 	exist, err := l.svcCtx.ItickTenantCategoryModel.FindOneByTenantIdCategoryId(l.ctx, in.TenantId, category.Id)
	// 	if err != nil && !errors.Is(err, models.ErrNotFound) {
	// 		return nil, err
	// 	}

	// 	if exist == nil {
	// 		_, err = l.svcCtx.ItickTenantCategoryModel.Insert(l.ctx, &models.TItickTenantCategory{
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
	// 		if err := l.svcCtx.ItickTenantCategoryModel.Update(l.ctx, exist); err != nil {
	// 			return nil, err
	// 		}
	// 		categoryCount++
	// 	}
	// }

	// products, err := collectProducts(l.ctx, l.svcCtx.ItickProductModel)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, product := range products {
	// 	exist, err := l.svcCtx.ItickTenantProductModel.FindOneByTenantIdProductId(l.ctx, in.TenantId, product.Id)
	// 	if err != nil && !errors.Is(err, models.ErrNotFound) {
	// 		return nil, err
	// 	}

	// 	if exist == nil {
	// 		_, err = l.svcCtx.ItickTenantProductModel.Insert(l.ctx, &models.TItickTenantProduct{
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
	// 		if err := l.svcCtx.ItickTenantProductModel.Update(l.ctx, exist); err != nil {
	// 			return nil, err
	// 		}
	// 		productCount++
	// 	}
	// }

	return &itick.InitTenantItickDisplayResp{
		Base:          helper.OkResp(),
		CategoryCount: categoryCount,
		ProductCount:  productCount,
	}, nil
}
