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

type ListWithdrawOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWithdrawOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawOrdersLogic {
	return &ListWithdrawOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现订单列表
func (l *ListWithdrawOrdersLogic) ListWithdrawOrders(in *payment.ListWithdrawOrdersReq) (*payment.ListWithdrawOrdersResp, error) {
	orders, total, err := l.svcCtx.WithdrawOrderModel.FindPage(
		l.ctx,
		in.TenantId,
		in.UserId,
		in.OrderNo,
		0,
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(orders)) == in.Page.Limit {
		lastItem := orders[len(orders)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(orders)) == in.Page.Limit

	data := make([]*payment.WithdrawOrder, 0, len(orders))
	for _, o := range orders {
		data = append(data, &payment.WithdrawOrder{
			Id:           o.Id,
			TenantId:     o.TenantId,
			UserId:       o.UserId,
			OrderNo:      o.OrderNo,
			BizOrderNo:   o.BizOrderNo.String,
			Currency:     o.Currency,
			Amount:       o.Amount,
			FeeAmount:    o.FeeAmount,
			ActualAmount: o.ActualAmount,
			ClientType:   payment.ClientType(o.ClientType),
			ClientIp:     o.ClientIp.String,
			Status:       payment.PayOrderStatus(o.Status),
			ThirdTradeNo: o.ThirdTradeNo.String,
			ThirdOrderNo: o.ThirdOrderNo.String,
			RequestData:  o.RequestData.String,
			ResponseData: o.ResponseData.String,
			NotifyData:   o.NotifyData.String,
			ProcessTime:  o.ProcessTime.Int64,
			NotifyTime:   o.NotifyTime.Int64,
			CloseTime:    o.CloseTime.Int64,
			Remark:       o.Remark.String,
			CreateTimes:  o.CreateTimes,
			UpdateTimes:  o.UpdateTimes,
		})
	}

	return &payment.ListWithdrawOrdersResp{
		Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor),
		Data: data,
	}, nil
}
