package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryNotifyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryNotifyLogic {
	return &RetryNotifyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 重试回调
func (l *RetryNotifyLogic) RetryNotify(in *payment.RetryNotifyReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "RetryNotify"
	)

	// 查询订单是否存在
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if order == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, "订单不存在"),
		}, nil
	}

	// 只有已支付的订单才需要重试回调
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS) {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(201, "只有已支付订单才能重试回调"),
		}, nil
	}

	// 查询回调日志
	notifyLog, err := l.svcCtx.RechargeNotifyLogModel.FindOneByOrderId(l.ctx, order.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if notifyLog == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, "回调记录不存在"),
		}, nil
	}

	// 更新回调状态，标记为待重试
	notifyLog.NotifyStatus = int64(payment.NotifyProcessStatus_NOTIFY_PROCESS_STATUS_PENDING)

	err = l.svcCtx.RechargeNotifyLogModel.Update(l.ctx, notifyLog)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Retry notify success: %s", in.OrderNo)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
