package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyRechargeOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMyRechargeOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyRechargeOrderStatusLogic {
	return &QueryMyRechargeOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 轮询订单状态
func (l *QueryMyRechargeOrderStatusLogic) QueryMyRechargeOrderStatus(in *payment.QueryMyRechargeOrderStatusReq) (*payment.QueryMyRechargeOrderStatusResp, error) {
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if order == nil {
		return &payment.QueryMyRechargeOrderStatusResp{
			Base: &common.RespBase{
				Code: 404,
				Msg:  "订单不存在",
			},
		}, nil
	}

	// Check permission
	if order.UserId != in.UserId || order.TenantId != in.TenantId {
		return &payment.QueryMyRechargeOrderStatusResp{
			Base: &common.RespBase{
				Code: 403,
				Msg:  "无权查询该订单",
			},
		}, nil
	}

	return &payment.QueryMyRechargeOrderStatusResp{
		Base: helper.OkResp(),
		Order: &payment.RechargeOrder{
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
