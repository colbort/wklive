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

type CreateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 产品
func (l *CreateProductLogic) CreateProduct(in *market.CreateProductReq) (*market.CommonResp, error) {
	category, err := l.svcCtx.MarketCategoryModel.FindOneByCategoryType(l.ctx, int64(in.CategoryType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if category == nil {
		return &market.CommonResp{
			Base: helper.ErrResp(i18n.CategoryNotFound, i18n.Translate(i18n.CategoryNotFound, l.ctx)),
		}, nil
	}

	exist, err := l.svcCtx.MarketProductModel.FindOneByCategoryTypeMarketSymbol(l.ctx, int64(in.CategoryType), in.Market, in.Symbol)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return &market.CommonResp{
			Base: helper.ErrResp(i18n.ResourceAlreadyExists, i18n.Translate(i18n.ResourceAlreadyExists, l.ctx)),
		}, nil
	}

	now := cutils.NowMillis()
	_, err = l.svcCtx.MarketProductModel.Insert(l.ctx, &models.TItickProduct{
		CategoryType: int64(in.CategoryType),
		CategoryName: category.CategoryName,
		CategoryCode: category.CategoryCode,
		Market:       in.Market,
		Symbol:       in.Symbol,
		Code:         in.Code,
		Name:         in.Name,
		DisplayName:  in.DisplayName,
		BaseCoin:     in.BaseCoin,
		QuoteCoin:    in.QuoteCoin,
		Enabled:      int64(in.Enabled),
		AppVisible:   int64(in.AppVisible),
		Sort:         in.Sort,
		Icon:         in.Icon,
		SyncPriority: int64(in.SyncPriority),
		Remark:       in.Remark,
		CreateTimes:  now,
		UpdateTimes:  now,
	})
	if err != nil {
		return nil, err
	}

	return &market.CommonResp{Base: helper.OkResp()}, nil
}
