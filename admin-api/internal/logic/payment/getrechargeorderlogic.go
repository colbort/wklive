// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRechargeOrderLogic {
	return &GetRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRechargeOrderLogic) GetRechargeOrder(req *types.GetRechargeOrderReq) (resp *types.GetRechargeOrderResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetRechargeOrder(l.ctx, &payment.GetRechargeOrderReq{
		TenantId: req.TenantId,
		OrderNo:  req.OrderNo,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetRechargeOrderResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.RechargeOrder{
			Id:           result.Data.Id,
			TenantId:     result.Data.TenantId,
			UserId:       result.Data.UserId,
			OrderNo:      result.Data.OrderNo,
			BizOrderNo:   result.Data.BizOrderNo,
			PlatformId:   result.Data.PlatformId,
			ProductId:    result.Data.ProductId,
			AccountId:    result.Data.AccountId,
			ChannelId:    result.Data.ChannelId,
			Currency:     result.Data.Currency,
			OrderAmount:  result.Data.OrderAmount,
			PayAmount:    result.Data.PayAmount,
			FeeAmount:    result.Data.FeeAmount,
			Subject:      result.Data.Subject,
			Body:         result.Data.Body,
			ClientType:   int64(result.Data.ClientType),
			ClientIp:     result.Data.ClientIp,
			Status:       int64(result.Data.Status),
			ThirdTradeNo: result.Data.ThirdTradeNo,
			ThirdOrderNo: result.Data.ThirdOrderNo,
			PayUrl:       result.Data.PayUrl,
			QrContent:    result.Data.QrContent,
			RequestData:  result.Data.RequestData,
			ResponseData: result.Data.ResponseData,
			NotifyData:   result.Data.NotifyData,
			ExpireTime:   result.Data.ExpireTime,
			PaidTime:     result.Data.PaidTime,
			NotifyTime:   result.Data.NotifyTime,
			CloseTime:    result.Data.CloseTime,
			Remark:       result.Data.Remark,
			CreateTimes:  result.Data.CreateTimes,
			UpdateTimes:  result.Data.UpdateTimes,
		},
	}
	return resp, nil
}
