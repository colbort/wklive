package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

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
func (l *BatchUpsertTenantProductsLogic) BatchUpsertTenantProducts(in *itick.BatchUpsertTenantProductsReq) (*itick.AdminCommonResp, error) {
	existing, err := collectTenantProducts(l.ctx, l.svcCtx.ItickTenantProductModel, in.TenantId)
	if err != nil {
		return nil, err
	}

	existingByID := make(map[int64]*models.TItickTenantProduct, len(existing))
	existingByProductID := make(map[int64]*models.TItickTenantProduct, len(existing))
	for _, item := range existing {
		existingByID[item.Id] = item
		existingByProductID[item.ProductId] = item
	}

	keptIDs := make(map[int64]struct{}, len(in.Data))
	now := cutils.NowMillis()
	for _, item := range in.Data {
		if item.Id > 0 {
			exist := existingByID[item.Id]
			if exist == nil {
				return &itick.AdminCommonResp{
					Base: helper.GetErrResp(1, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
				}, nil
			}
			exist.Enabled = item.Enabled
			exist.AppVisible = item.AppVisible
			exist.Sort = item.Sort
			exist.Remark = item.Remark
			exist.UpdateTimes = now
			if err := l.svcCtx.ItickTenantProductModel.Update(l.ctx, exist); err != nil {
				return nil, err
			}
			keptIDs[exist.Id] = struct{}{}
			continue
		}

		if _, err := l.svcCtx.ItickProductModel.FindOne(l.ctx, item.ProductId); err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}

		if exist := existingByProductID[item.ProductId]; exist != nil {
			exist.Enabled = item.Enabled
			exist.AppVisible = item.AppVisible
			exist.Sort = item.Sort
			exist.Remark = item.Remark
			exist.UpdateTimes = now
			if err := l.svcCtx.ItickTenantProductModel.Update(l.ctx, exist); err != nil {
				return nil, err
			}
			keptIDs[exist.Id] = struct{}{}
			continue
		}

		result, err := l.svcCtx.ItickTenantProductModel.Insert(l.ctx, &models.TItickTenantProduct{
			TenantId:    in.TenantId,
			ProductId:   item.ProductId,
			Enabled:     item.Enabled,
			AppVisible:  item.AppVisible,
			Sort:        item.Sort,
			Remark:      item.Remark,
			CreateTimes: now,
			UpdateTimes: now,
		})
		if err != nil {
			return nil, err
		}
		if id, err := result.LastInsertId(); err == nil {
			keptIDs[id] = struct{}{}
		}
	}

	for _, item := range existing {
		if _, ok := keptIDs[item.Id]; ok {
			continue
		}
		if err := l.svcCtx.ItickTenantProductModel.Delete(l.ctx, item.Id); err != nil {
			return nil, err
		}
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
