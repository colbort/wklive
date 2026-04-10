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

type GetMyWithdrawOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyWithdrawOrderLogic {
	return &GetMyWithdrawOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现订单详情
func (l *GetMyWithdrawOrderLogic) GetMyWithdrawOrder(in *payment.GetMyWithdrawOrderReq) (*payment.GetMyWithdrawOrderResp, error) {
	order, err := l.svcCtx.WithdrawOrderModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if order == nil {
		return &payment.GetMyWithdrawOrderResp{
			Base: helper.GetErrResp(404, "订单不存在"),
		}, nil
	}

	// Check permission
	if order.UserId != in.UserId {
		return &payment.GetMyWithdrawOrderResp{
			Base: helper.GetErrResp(403, "无权访问该订单"),
		}, nil
	}

	return &payment.GetMyWithdrawOrderResp{
		Base: helper.OkResp(),
		Data: &payment.WithdrawOrder{
			Id:          order.Id,
			TenantId:    order.TenantId,
			UserId:      order.UserId,
			OrderNo:     order.OrderNo,
			Amount:      order.Amount,
			Currency:    order.Currency,
			Status:      payment.PayOrderStatus(order.Status),
			Remark:      order.Remark.String,
			CreateTimes: order.CreateTimes,
			UpdateTimes: order.UpdateTimes,
		},
	}, nil
}
