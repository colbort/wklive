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

type GetMyRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyRechargeOrderLogic {
	return &GetMyRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyRechargeOrderLogic) GetMyRechargeOrder(req *types.GetMyRechargeOrderReq) (resp *types.GetMyRechargeOrderResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	tenantId := req.TenantId
	if tenantId == 0 {
		tenantId, err = utils.GetTenantIdFromCtx(l.ctx)
		if err != nil {
			return nil, err
		}
	}

	result, err := l.svcCtx.PaymentCli.GetMyRechargeOrder(l.ctx, &payment.GetMyRechargeOrderReq{
		TenantId: tenantId,
		UserId:   userId,
		OrderNo:  req.OrderNo,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetMyRechargeOrderResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.RechargeOrder{
			Id:           result.Order.Id,
			TenantId:     result.Order.TenantId,
			UserId:       result.Order.UserId,
			OrderNo:      result.Order.OrderNo,
			BizOrderNo:   result.Order.BizOrderNo,
			PlatformId:   result.Order.PlatformId,
			ProductId:    result.Order.ProductId,
			AccountId:    result.Order.AccountId,
			ChannelId:    result.Order.ChannelId,
			Currency:     result.Order.Currency,
			OrderAmount:  result.Order.OrderAmount,
			PayAmount:    result.Order.PayAmount,
			FeeAmount:    result.Order.FeeAmount,
			Subject:      result.Order.Subject,
			Body:         result.Order.Body,
			ClientType:   int64(result.Order.ClientType),
			ClientIp:     result.Order.ClientIp,
			Status:       int64(result.Order.Status),
			ThirdTradeNo: result.Order.ThirdTradeNo,
			ThirdOrderNo: result.Order.ThirdOrderNo,
			PayUrl:       result.Order.PayUrl,
			QrContent:    result.Order.QrContent,
			RequestData:  result.Order.RequestData,
			ResponseData: result.Order.ResponseData,
			NotifyData:   result.Order.NotifyData,
			ExpireTime:   result.Order.ExpireTime,
			PaidTime:     result.Order.PaidTime,
			NotifyTime:   result.Order.NotifyTime,
			CloseTime:    result.Order.CloseTime,
			Remark:       result.Order.Remark,
			CreateTime:   result.Order.CreateTime,
			UpdateTime:   result.Order.UpdateTime,
		},
	}

	return
}
