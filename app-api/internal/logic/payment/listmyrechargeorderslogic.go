// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyRechargeOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyRechargeOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyRechargeOrdersLogic {
	return &ListMyRechargeOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyRechargeOrdersLogic) ListMyRechargeOrders(req *types.ListMyRechargeOrdersReq) (resp *types.ListMyRechargeOrdersResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.PaymentCli.ListMyRechargeOrders(l.ctx, &payment.ListMyRechargeOrdersReq{
		UserId:   userId,
		TenantId: tenantId,
		Page: &payment.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Status:          payment.PayOrderStatus(req.Status),
		OrderNo:         req.OrderNo,
		CreateTimeStart: req.CreateTimeStart,
		CreateTimeEnd:   req.CreateTimeEnd,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.RechargeOrder, 0)
	for _, order := range result.List {
		data = append(data, types.RechargeOrder{
			Id:           order.Id,
			TenantId:     order.TenantId,
			UserId:       order.UserId,
			OrderNo:      order.OrderNo,
			BizOrderNo:   order.BizOrderNo,
			PlatformId:   order.PlatformId,
			ProductId:    order.ProductId,
			AccountId:    order.AccountId,
			ChannelId:    order.ChannelId,
			Currency:     order.Currency,
			OrderAmount:  order.OrderAmount,
			PayAmount:    order.PayAmount,
			FeeAmount:    order.FeeAmount,
			Subject:      order.Subject,
			Body:         order.Body,
			ClientType:   int64(order.ClientType.Number()),
			ClientIp:     order.ClientIp,
			Status:       int64(order.Status.Number()),
			ThirdTradeNo: order.ThirdTradeNo,
			ThirdOrderNo: order.ThirdOrderNo,
			PayUrl:       order.PayUrl,
			QrContent:    order.QrContent,
			RequestData:  order.RequestData,
			ResponseData: order.ResponseData,
			NotifyData:   order.NotifyData,
			ExpireTime:   order.ExpireTime,
			PaidTime:     order.PaidTime,
			NotifyTime:   order.NotifyTime,
			CloseTime:    order.CloseTime,
			Remark:       order.Remark,
			CreateTime:   order.CreateTime,
			UpdateTime:   order.UpdateTime,
		})
	}
	return &types.ListMyRechargeOrdersResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}, nil
}
