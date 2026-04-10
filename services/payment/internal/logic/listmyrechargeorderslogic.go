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

type ListMyRechargeOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyRechargeOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyRechargeOrdersLogic {
	return &ListMyRechargeOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的充值订单列表
func (l *ListMyRechargeOrdersLogic) ListMyRechargeOrders(in *payment.ListMyRechargeOrdersReq) (*payment.ListMyRechargeOrdersResp, error) {
	items, total, err := l.svcCtx.RechargeOrderModel.FindPage(l.ctx, in.TenantId, in.UserId, in.OrderNo, int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit

	data := make([]*payment.RechargeOrder, 0)
	for _, order := range items {
		data = append(data, &payment.RechargeOrder{
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
		})
	}

	return &payment.ListMyRechargeOrdersResp{
		Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor),
		List: data,
	}, nil
}
