package adminlogic

import (
	"context"
	"fmt"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchUpsertTenantProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchUpsertTenantProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpsertTenantProductsLogic {
	return &BatchUpsertTenantProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量更新租户产品，已关联的修改状态、排序和备注，未关联的新增，未提交的删除
func (l *BatchUpsertTenantProductsLogic) BatchUpsertTenantProducts(in *market.BatchUpsertTenantProductsReq) (*market.CommonResp, error) {
	if base, err := helpers.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &market.CommonResp{Base: base}, nil
	}
	requested := make(map[int64]struct{}, len(in.Data))
	for _, input := range in.Data {
		if input == nil || input.ProductId <= 0 {
			continue
		}
		if _, duplicate := requested[input.ProductId]; duplicate {
			return nil, fmt.Errorf("duplicate product_id: %d", input.ProductId)
		}
		requested[input.ProductId] = struct{}{}
		product, err := l.svcCtx.MarketProductModel.FindOne(l.ctx, input.ProductId)
		if err != nil {
			return nil, err
		}
		if base := validateSelectableProduct(l.ctx, product); base != nil {
			return &market.CommonResp{Base: base}, nil
		}
	}

	current := make([]*models.TItickTenantProduct, 0)
	var cursor int64
	for {
		page, _, err := l.svcCtx.MarketTenantProductModel.FindPage(l.ctx, models.TenantProductPageFilter{
			TenantId: in.TenantId,
		}, cursor, 100)
		if err != nil {
			return nil, err
		}
		current = append(current, page...)
		if len(page) < 100 {
			break
		}
		cursor = page[len(page)-1].Id
	}
	byProduct := make(map[int64]*models.TItickTenantProduct, len(current))
	for _, item := range current {
		byProduct[item.ProductId] = item
	}
	now := cutils.NowMillis()
	for _, input := range in.Data {
		if input == nil || input.ProductId <= 0 {
			continue
		}
		item := byProduct[input.ProductId]
		if item == nil {
			_, err := l.svcCtx.MarketTenantProductModel.Insert(l.ctx, &models.TItickTenantProduct{
				TenantId: in.TenantId, ProductId: input.ProductId, Enabled: int64(input.Enabled),
				AppVisible: int64(input.AppVisible), Sort: input.Sort, Remark: input.Remark,
				CreateTimes: now, UpdateTimes: now,
			})
			if err != nil {
				return nil, err
			}
		} else {
			item.Enabled = int64(input.Enabled)
			item.AppVisible = int64(input.AppVisible)
			item.Sort = input.Sort
			item.Remark = input.Remark
			item.UpdateTimes = now
			err := l.svcCtx.MarketTenantProductModel.Update(l.ctx, item)
			if err != nil {
				return nil, err
			}
			continue
		}
	}
	for _, item := range current {
		if _, keep := requested[item.ProductId]; keep {
			continue
		}
		if err := l.svcCtx.MarketTenantProductModel.Delete(l.ctx, item.Id); err != nil {
			return nil, err
		}
	}
	if err := l.svcCtx.MarketManager.RefreshActiveProductSubscriptions(l.ctx); err != nil {
		l.Errorf("refresh active products after batch update failed: %v", err)
	}

	return &market.CommonResp{Base: helper.OkResp()}, nil
}
