package adminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantCategoryLogic {
	return &CreateTenantCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户产品类型
func (l *CreateTenantCategoryLogic) CreateTenantCategory(in *market.CreateTenantCategoryReq) (*market.CommonResp, error) {
	category, err := l.svcCtx.MarketCategoryModel.FindOne(l.ctx, in.CategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return &market.CommonResp{
			Base: helper.ErrResp(i18n.CategoryNotFound, i18n.Translate(i18n.CategoryNotFound, l.ctx)),
		}, nil
	}

	exist, err := l.svcCtx.MarketTenantCategoryModel.FindOneByTenantIdCategoryId(l.ctx, in.TenantId, in.CategoryId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return &market.CommonResp{
			Base: helper.ErrResp(i18n.ResourceAlreadyExists, i18n.Translate(i18n.ResourceAlreadyExists, l.ctx)),
		}, nil
	}

	now := cutils.NowMillis()
	_, err = l.svcCtx.MarketTenantCategoryModel.Insert(l.ctx, &models.TItickTenantCategory{
		TenantId:    in.TenantId,
		CategoryId:  in.CategoryId,
		Enabled:     int64(in.Enabled),
		AppVisible:  int64(in.AppVisible),
		Sort:        in.Sort,
		Remark:      in.Remark,
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		return nil, err
	}

	return &market.CommonResp{Base: helper.OkResp()}, nil
}
