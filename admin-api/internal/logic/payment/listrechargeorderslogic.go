// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRechargeOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRechargeOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRechargeOrdersLogic {
	return &ListRechargeOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRechargeOrdersLogic) ListRechargeOrders(req *types.ListRechargeOrdersReq) (resp *types.ListRechargeOrdersResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListRechargeOrders(l.ctx, &payment.ListRechargeOrdersReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId:        req.TenantId,
		UserId:          req.UserId,
		PlatformId:      req.PlatformId,
		ProductId:       req.ProductId,
		AccountId:       req.AccountId,
		ChannelId:       req.ChannelId,
		OrderNo:         req.OrderNo,
		BizOrderNo:      req.BizOrderNo,
		ThirdTradeNo:    req.ThirdTradeNo,
		Status:          payment.PayOrderStatus(req.Status),
		CreateTimeStart: req.CreateTimeStart,
		CreateTimeEnd:   req.CreateTimeEnd,
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.RechargeOrder, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.RechargeOrder{
			Id:           item.Id,
			TenantId:     item.TenantId,
			UserId:       item.UserId,
			OrderNo:      item.OrderNo,
			BizOrderNo:   item.BizOrderNo,
			PlatformId:   item.PlatformId,
			ProductId:    item.ProductId,
			AccountId:    item.AccountId,
			ChannelId:    item.ChannelId,
			Currency:     item.Currency,
			OrderAmount:  item.OrderAmount,
			PayAmount:    item.PayAmount,
			FeeAmount:    item.FeeAmount,
			Subject:      item.Subject,
			Body:         item.Body,
			ClientType:   int64(item.ClientType),
			ClientIp:     item.ClientIp,
			Status:       int64(item.Status),
			ThirdTradeNo: item.ThirdTradeNo,
			ThirdOrderNo: item.ThirdOrderNo,
			PayUrl:       item.PayUrl,
			QrContent:    item.QrContent,
			RequestData:  item.RequestData,
			ResponseData: item.ResponseData,
			NotifyData:   item.NotifyData,
			ExpireTime:   item.ExpireTime,
			PaidTime:     item.PaidTime,
			NotifyTime:   item.NotifyTime,
			CloseTime:    item.CloseTime,
			Remark:       item.Remark,
			CreateTimes:   item.CreateTimes,
			UpdateTimes:   item.UpdateTimes,
		}
	}

	resp = &types.ListRechargeOrdersResp{
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
	}
	return resp, nil
}
