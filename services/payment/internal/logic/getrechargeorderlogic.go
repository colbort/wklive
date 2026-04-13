package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRechargeOrderLogic {
	return &GetRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取充值订单
func (l *GetRechargeOrderLogic) GetRechargeOrder(in *payment.GetRechargeOrderReq) (*payment.GetRechargeOrderResp, error) {
	var (
		errLogic = "GetRechargeOrder"
	)

	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if order == nil {
		return &payment.GetRechargeOrderResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetRechargeOrderResp{
		Base: helper.OkResp(),
		Data: toRechargeOrderProto(order),
	}, nil
}
