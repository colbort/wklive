package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
)

type CancelMyRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelMyRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelMyRechargeOrderLogic {
	return &CancelMyRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取消未支付订单
func (l *CancelMyRechargeOrderLogic) CancelMyRechargeOrder(in *payment.CancelMyRechargeOrderReq) (*payment.AppCommonResp, error) {
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if order == nil {
		return &payment.AppCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}

	// Check permission
	if order.UserId != in.UserId || order.TenantId != in.TenantId {
		return &payment.AppCommonResp{
			Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionCancelOrder, l.ctx)),
		}, nil
	}

	// Can only cancel unpaid orders
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) {
		return &payment.AppCommonResp{
			Base: helper.GetErrResp(201, i18n.Translate(i18n.OnlyPendingPaymentOrdersCanCancel, l.ctx)),
		}, nil
	}

	// Update order status to cancelled
	order.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_CLOSED)
	order.UpdateTimes = time.Now().UnixMilli()

	err = l.svcCtx.RechargeOrderModel.Update(l.ctx, order)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("Cancel recharge order success: %s, user_id: %d", in.OrderNo, in.UserId)

	return &payment.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
