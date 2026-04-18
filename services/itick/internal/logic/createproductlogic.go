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
func (l *CreateProductLogic) CreateProduct(in *itick.CreateProductReq) (*itick.AdminCommonResp, error) {
	category, err := l.svcCtx.ItickCategoryModel.FindOneByCategoryType(l.ctx, int64(in.CategoryType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if category == nil {
		return &itick.AdminCommonResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.CategoryNotFound, l.ctx)),
		}, nil
	}

	exist, err := l.svcCtx.ItickProductModel.FindOneByCategoryTypeMarketSymbol(l.ctx, int64(in.CategoryType), in.Market, in.Symbol)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return &itick.AdminCommonResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.ResourceAlreadyExists, l.ctx)),
		}, nil
	}

	now := cutils.NowMillis()
	_, err = l.svcCtx.ItickProductModel.Insert(l.ctx, &models.TItickProduct{
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
		Enabled:      in.Enabled,
		AppVisible:   in.AppVisible,
		Sort:         in.Sort,
		Icon:         in.Icon,
		Remark:       in.Remark,
		CreateTimes:  now,
		UpdateTimes:  now,
	})
	if err != nil {
		return nil, err
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
