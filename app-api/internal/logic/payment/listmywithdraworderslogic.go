// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyWithdrawOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyWithdrawOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyWithdrawOrdersLogic {
	return &ListMyWithdrawOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyWithdrawOrdersLogic) ListMyWithdrawOrders(req *types.ListMyWithdrawOrdersReq) (resp *types.ListMyWithdrawOrdersResp, err error) {
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

	result, err := l.svcCtx.PaymentCli.ListMyWithdrawOrders(l.ctx, &payment.ListMyWithdrawOrdersReq{
		TenantId: tenantId,
		UserId:   userId,
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Status: payment.PayOrderStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	resp = &types.ListMyWithdrawOrdersResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: make([]types.WithdrawOrder, 0, len(result.Data)),
	}
	for _, item := range result.Data {
		resp.Data = append(resp.Data, types.WithdrawOrder{
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
			CreateTimes:  item.CreateTimes,
			UpdateTimes:  item.UpdateTimes,
		})
	}

	return
}
