package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
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
			Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}

	// 只有pending状态的订单才能标记为成功
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(201, i18n.Translate(i18n.OnlyPendingPaymentOrdersCanMarkSuccess, l.ctx)),
		}, nil
	}

	now := time.Now().UnixMilli()
	order.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS)
	order.PayAmount = in.PayAmount
	order.ThirdTradeNo = sql.NullString{String: in.ThirdTradeNo, Valid: in.ThirdTradeNo != ""}
	order.UpdateTimes = now

	err = l.svcCtx.RechargeOrderModel.Update(l.ctx, order)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Manual mark recharge order success: %s", in.OrderNo)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
