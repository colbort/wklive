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
		Data: &payment.RechargeOrder{
			Id:           order.Id,
			TenantId:     order.TenantId,
			UserId:       order.UserId,
			OrderNo:      order.OrderNo,
			BizOrderNo:   order.BizOrderNo.String,
			PlatformId:   order.PlatformId,
			ProductId:    order.ProductId,
			AccountId:    order.AccountId,
			ChannelId:    order.ChannelId,
			Currency:     order.Currency,
			OrderAmount:  order.OrderAmount,
			PayAmount:    order.PayAmount,
			FeeAmount:    order.FeeAmount,
			Subject:      order.Subject.String,
			Body:         order.Body.String,
			ClientType:   payment.ClientType(order.ClientType),
			ClientIp:     order.ClientIp.String,
			Status:       payment.PayOrderStatus(order.Status),
			ThirdTradeNo: order.ThirdTradeNo.String,
			CreateTimes:  order.CreateTimes,
			UpdateTimes:  order.UpdateTimes,
		},
	}, nil
}
