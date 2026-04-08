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

type ListWithdrawOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListWithdrawOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawOrdersLogic {
	return &ListWithdrawOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListWithdrawOrdersLogic) ListWithdrawOrders(req *types.ListWithdrawOrdersReq) (resp *types.ListWithdrawOrdersResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListWithdrawOrders(l.ctx, &payment.ListWithdrawOrdersReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId: req.TenantId,
		UserId:   req.UserId,
		OrderNo:  req.OrderNo,
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.WithdrawOrder, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.WithdrawOrder{
			Id:           item.Id,
			TenantId:     item.TenantId,
			UserId:       item.UserId,
			OrderNo:      item.OrderNo,
			BizOrderNo:   item.BizOrderNo,
			Currency:     item.Currency,
			Amount:       item.Amount,
			FeeAmount:    item.FeeAmount,
			ActualAmount: item.ActualAmount,
			ClientType:   int64(item.ClientType),
			ClientIp:     item.ClientIp,
			Status:       int64(item.Status),
			ThirdTradeNo: item.ThirdTradeNo,
			ThirdOrderNo: item.ThirdOrderNo,
			RequestData:  item.RequestData,
			ResponseData: item.ResponseData,
			NotifyData:   item.NotifyData,
			ProcessTime:  item.ProcessTime,
			NotifyTime:   item.NotifyTime,
			CloseTime:    item.CloseTime,
			Remark:       item.Remark,
			CreateTimes:   item.CreateTimes,
			UpdateTimes:   item.UpdateTimes,
		}
	}

	return &types.ListWithdrawOrdersResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
