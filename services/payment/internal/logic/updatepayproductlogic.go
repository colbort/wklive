package logic

import (
	"context"
	"database/sql"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
)

type UpdatePayProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePayProductLogic {
	return &UpdatePayProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新产品
func (l *UpdatePayProductLogic) UpdatePayProduct(in *payment.UpdatePayProductReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "UpdatePayProduct"
	)

	// 査询产品是否存在
	product, err := l.svcCtx.PayProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if product == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx)),
		}, nil
	}

	now := time.Now().UnixMilli()
	if in.ProductName != "" {
		product.ProductName = in.ProductName
	}
	if in.SceneType != 0 {
		product.SceneType = int64(in.SceneType)
	}
	if in.Currency != "" {
		product.Currency = in.Currency
	}
	if in.Status != 0 {
		product.Status = int64(in.Status)
	}
	if in.Remark != "" {
		product.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	product.UpdateTimes = now

	err = l.svcCtx.PayProductModel.Update(l.ctx, product)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Update pay product success: %d", in.Id)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
