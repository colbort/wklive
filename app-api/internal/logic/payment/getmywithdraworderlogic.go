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

type GetMyWithdrawOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyWithdrawOrderLogic {
	return &GetMyWithdrawOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyWithdrawOrderLogic) GetMyWithdrawOrder(req *types.GetMyWithdrawOrderReq) (resp *types.GetMyWithdrawOrderResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.PaymentCli.GetMyWithdrawOrder(l.ctx, &payment.GetMyWithdrawOrderReq{
		Id:     req.Id,
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.GetMyWithdrawOrderResp{
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
	}

	return
}
