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

type BatchUpsertTenantCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchUpsertTenantCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpsertTenantCategoriesLogic {
	return &BatchUpsertTenantCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量更新租户产品类型，已关联的修改状态、排序和备注，未关联的新增，未提交的删除
func (l *BatchUpsertTenantCategoriesLogic) BatchUpsertTenantCategories(in *itick.BatchUpsertTenantCategoriesReq) (*itick.AdminCommonResp, error) {
	existing, err := collectTenantCategories(l.ctx, l.svcCtx.ItickTenantCategoryModel, in.TenantId)
	if err != nil {
		return nil, err
	}

	existingByID := make(map[int64]*models.TItickTenantCategory, len(existing))
	existingByCategoryID := make(map[int64]*models.TItickTenantCategory, len(existing))
	for _, item := range existing {
		existingByID[item.Id] = item
		existingByCategoryID[item.CategoryId] = item
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
			if err := l.svcCtx.ItickTenantCategoryModel.Update(l.ctx, exist); err != nil {
				return nil, err
			}
			keptIDs[exist.Id] = struct{}{}
			continue
		}

		if _, err := l.svcCtx.ItickCategoryModel.FindOne(l.ctx, item.CategoryId); err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}

		if exist := existingByCategoryID[item.CategoryId]; exist != nil {
			exist.Enabled = item.Enabled
			exist.AppVisible = item.AppVisible
			exist.Sort = item.Sort
			exist.Remark = item.Remark
			exist.UpdateTimes = now
			if err := l.svcCtx.ItickTenantCategoryModel.Update(l.ctx, exist); err != nil {
				return nil, err
			}
			keptIDs[exist.Id] = struct{}{}
			continue
		}

		result, err := l.svcCtx.ItickTenantCategoryModel.Insert(l.ctx, &models.TItickTenantCategory{
			TenantId:    in.TenantId,
			CategoryId:  item.CategoryId,
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
		if err := l.svcCtx.ItickTenantCategoryModel.Delete(l.ctx, item.Id); err != nil {
			return nil, err
		}
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
