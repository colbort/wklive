package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CloseRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseRechargeOrderLogic {
	return &CloseRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 关闭充值订单
func (l *CloseRechargeOrderLogic) CloseRechargeOrder(in *payment.CloseRechargeOrderReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "CloseRechargeOrder"
	)

	// 査询订单是否存在
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if order == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}

	// 仅有pending状态的订单才能关闭
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(201, i18n.Translate(i18n.OnlyUnpaidOrdersCanClose, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	order.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_CLOSED)
	order.UpdateTimes = now

	err = l.svcCtx.RechargeOrderModel.Update(l.ctx, order)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Close recharge order success: %s", in.OrderNo)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
