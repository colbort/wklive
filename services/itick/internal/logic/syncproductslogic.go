package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncProductsLogic {
	return &SyncProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步产品列表 （定时任务）
func (l *SyncProductsLogic) SyncProducts(in *itick.SyncProductsReq) (*itick.SyncProductsResp, error) {
	categories, err := l.svcCtx.ItickCategoryModel.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}

	worker := NewSyncCategoryProductsWorker(l.ctx, l.svcCtx)
	for _, category := range categories {
		if category == nil {
			continue
		}
		if _, err := l.svcCtx.ItickCategoryModel.FindOne(l.ctx, category.Id); err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if err := worker.doSync(&itick.SyncCategoryProductsReq{Id: category.Id}); err != nil {
			return &itick.SyncProductsResp{
				Base: helper.GetErrResp(1, err.Error()),
			}, nil
		}
	}

	return &itick.SyncProductsResp{
		Base: helper.OkResp(),
	}, nil
}
