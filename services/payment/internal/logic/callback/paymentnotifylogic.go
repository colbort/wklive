package callbacklogic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PaymentNotifyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPaymentNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentNotifyLogic {
	return &PaymentNotifyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PaymentNotifyLogic) PaymentNotify(in *payment.ThirdPartyNotifyReq) (*payment.ThirdPartyNotifyResp, error) {
	account, adapter, err := resolveNotifyAdapter(l.ctx, l.svcCtx, in)
	if err != nil {
		return notifyResponse(in.GetPlatformCode(), false), nil
	}
	result, err := adapter.PaymentNotify(l.ctx, account, toNotifyRequest(in))
	if err != nil || result == nil {
		l.Errorf("payment notification rejected, platform=%s account=%s err=%v", in.PlatformCode, in.AccountCode, err)
		return notifyResponse(in.PlatformCode, false), nil
	}
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, result.OrderNo)
	if err != nil {
		if !isNotFound(err) {
			l.Errorf("find recharge order for notification failed, orderNo=%s err=%v", result.OrderNo, err)
		}
		return notifyResponse(in.PlatformCode, false), nil
	}
	if order.TenantId != account.TenantId || order.AccountId != account.Id {
		return notifyResponse(in.PlatformCode, false), nil
	}
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS) {
		l.Errorf("payment notification settlement is not complete, orderNo=%s", order.OrderNo)
		return notifyResponse(in.PlatformCode, false), nil
	}
	return notifyResponse(in.PlatformCode, true), nil
}
