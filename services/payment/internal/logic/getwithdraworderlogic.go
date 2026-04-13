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

type GetWithdrawOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawOrderLogic {
	return &GetWithdrawOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现订单详情
func (l *GetWithdrawOrderLogic) GetWithdrawOrder(in *payment.GetWithdrawOrderReq) (*payment.GetWithdrawOrderResp, error) {
	var (
		errLogic = "GetWithdrawOrder"
	)

	order, err := l.svcCtx.WithdrawOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if order == nil {
		return &payment.GetWithdrawOrderResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetWithdrawOrderResp{
		Base: helper.OkResp(),
		Data: &payment.WithdrawOrder{
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
			Amount:       order.Amount,
			FeeAmount:    order.FeeAmount,
			ActualAmount: order.ActualAmount,
			ClientType:   payment.ClientType(order.ClientType),
			ClientIp:     order.ClientIp.String,
			Status:       payment.PayOrderStatus(order.Status),
			ThirdTradeNo: order.ThirdTradeNo.String,
			ThirdOrderNo: order.ThirdOrderNo.String,
			RequestData:  order.RequestData.String,
			ResponseData: order.ResponseData.String,
			NotifyData:   order.NotifyData.String,
			ProcessTime:  order.ProcessTime.Int64,
			NotifyTime:   order.NotifyTime.Int64,
			CloseTime:    order.CloseTime.Int64,
			Remark:       order.Remark.String,
			CreateTimes:  order.CreateTimes,
			UpdateTimes:  order.UpdateTimes,
		},
	}, nil
}
