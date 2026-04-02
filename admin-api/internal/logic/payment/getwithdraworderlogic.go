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

type GetWithdrawOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawOrderLogic {
	return &GetWithdrawOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetWithdrawOrderLogic) GetWithdrawOrder(req *types.GetWithdrawOrderReq) (resp *types.GetWithdrawOrderResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetWithdrawOrder(l.ctx, &payment.GetWithdrawOrderReq{
		TenantId: req.TenantId,
		OrderNo:  req.OrderNo,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetWithdrawOrderResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.WithdrawOrder{
			Id:           result.Data.Id,
			TenantId:     result.Data.TenantId,
			UserId:       result.Data.UserId,
			OrderNo:      result.Data.OrderNo,
			BizOrderNo:   result.Data.BizOrderNo,
			Currency:     result.Data.Currency,
			Amount:       result.Data.Amount,
			FeeAmount:    result.Data.FeeAmount,
			ActualAmount: result.Data.ActualAmount,
			ClientType:   int64(result.Data.ClientType),
			ClientIp:     result.Data.ClientIp,
			Status:       int64(result.Data.Status),
			ThirdTradeNo: result.Data.ThirdTradeNo,
			ThirdOrderNo: result.Data.ThirdOrderNo,
			RequestData:  result.Data.RequestData,
			ResponseData: result.Data.ResponseData,
			NotifyData:   result.Data.NotifyData,
			ProcessTime:  result.Data.ProcessTime,
			NotifyTime:   result.Data.NotifyTime,
			CloseTime:    result.Data.CloseTime,
			Remark:       result.Data.Remark,
			CreateTime:   result.Data.CreateTime,
			UpdateTime:   result.Data.UpdateTime,
		},
	}, nil
}
