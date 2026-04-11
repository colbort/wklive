package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
)

type GetPayProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayProductLogic {
	return &GetPayProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取产品详情
func (l *GetPayProductLogic) GetPayProduct(in *payment.GetPayProductReq) (*payment.GetPayProductResp, error) {
	var (
		errLogic = "GetPayProduct"
	)

	product, err := l.svcCtx.PayProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if product == nil {
		return &payment.GetPayProductResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetPayProductResp{
		Base: helper.OkResp(),
		Data: &payment.PayProduct{
			Id:          product.Id,
			PlatformId:  product.PlatformId,
			ProductCode: product.ProductCode,
			ProductName: product.ProductName,
			SceneType:   payment.SceneType(product.SceneType),
			Currency:    product.Currency,
			Status:      payment.CommonStatus(product.Status),
			Remark:      product.Remark.String,
			CreateTimes: product.CreateTimes,
			UpdateTimes: product.UpdateTimes,
		},
	}, nil
}
