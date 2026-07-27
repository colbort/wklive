package callbacklogic

import (
	"context"
	"wklive/services/payment/internal/logic/helpers"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayoutNotifyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPayoutNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayoutNotifyLogic {
	return &PayoutNotifyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PayoutNotifyLogic) PayoutNotify(in *payment.ThirdPartyNotifyReq) (*payment.ThirdPartyNotifyResp, error) {
	account, adapter, err := helpers.ResolveNotifyAdapter(l.ctx, l.svcCtx, in)
	if err != nil {
		return helpers.NotifyResponse(in.GetPlatformCode(), false), nil
	}
	result, err := adapter.PayoutNotify(l.ctx, account, helpers.ToNotifyRequest(in))
	if err != nil || result == nil {
		l.Errorf("payout notification rejected, platform=%s account=%s err=%v", in.PlatformCode, in.AccountCode, err)
		return helpers.NotifyResponse(in.PlatformCode, false), nil
	}
	order, err := l.svcCtx.WithdrawOrderModel.FindOneByOrderNo(l.ctx, result.OrderNo)
	if err != nil {
		if !helpers.IsNotFound(err) {
			l.Errorf("find withdraw order for notification failed, orderNo=%s err=%v", result.OrderNo, err)
		}
		return helpers.NotifyResponse(in.PlatformCode, false), nil
	}
	if order.TenantId != account.TenantId || order.AccountId != account.Id {
		return helpers.NotifyResponse(in.PlatformCode, false), nil
	}
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS) &&
		order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_FAILED) {
		l.Errorf("payout notification settlement is not complete, orderNo=%s", order.OrderNo)
		return helpers.NotifyResponse(in.PlatformCode, false), nil
	}
	return helpers.NotifyResponse(in.PlatformCode, true), nil
}
