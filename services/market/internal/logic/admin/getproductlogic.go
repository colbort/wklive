package adminlogic

import (
	"context"
	"errors"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取产品详情
func (l *GetProductLogic) GetProduct(in *market.GetProductReq) (*market.GetProductResp, error) {
	result, err := l.svcCtx.MarketProductModel.FindOne(l.ctx, in.Id)
	if err != nil && errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if result == nil {
		return &market.GetProductResp{
			Base: helper.ErrResp(i18n.NotFound, i18n.Translate(i18n.NotFound, l.ctx)),
		}, nil
	}
	return &market.GetProductResp{
		Base: helper.OkResp(),
		Data: helpers.ToProductProto(result),
	}, nil
}
