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

type ManualMarkRechargeOrderSuccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualMarkRechargeOrderSuccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualMarkRechargeOrderSuccessLogic {
	return &ManualMarkRechargeOrderSuccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 人工标记充值订单支付成功
func (l *ManualMarkRechargeOrderSuccessLogic) ManualMarkRechargeOrderSuccess(in *payment.ManualMarkRechargeOrderSuccessReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "ManualMarkRechargeOrderSuccess"
	)

	// 查询订单是否存在
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if order == nil {
		return &payment.AdminCommonResp{
			Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}
	if _, base, err := applyAdminTenantUpdateScope(l.ctx, order.TenantId, i18n.OrderNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return base, nil
	}

	// 只有待支付/支付中状态的订单才能标记为成功
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) &&
		order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PAYING) {
		return &payment.AdminCommonResp{
			Base: helper.ErrResp(i18n.OnlyPendingPaymentOrdersCanMarkSuccess, i18n.Translate(i18n.OnlyPendingPaymentOrdersCanMarkSuccess, l.ctx)),
		}, nil
	}

	if err := markRechargeOrderSuccessAndCredit(l.ctx, l.svcCtx, order, in.ThirdTradeNo, in.PayAmount, "manual mark recharge success"); err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Manual mark recharge order success: %s", in.OrderNo)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
