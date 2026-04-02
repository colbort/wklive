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

type ListUserRechargeStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserRechargeStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserRechargeStatsLogic {
	return &ListUserRechargeStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserRechargeStatsLogic) ListUserRechargeStats(req *types.ListUserRechargeStatsReq) (resp *types.ListUserRechargeStatsResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListUserRechargeStats(l.ctx, &payment.ListUserRechargeStatsReq{
		Page: &payment.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId:              req.TenantId,
		UserId:                req.UserId,
		SuccessTotalAmountMin: req.SuccessTotalAmountMin,
		SuccessTotalAmountMax: req.SuccessTotalAmountMax,
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.UserRechargeStat, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.UserRechargeStat{
			Id:                 item.Id,
			TenantId:           item.TenantId,
			UserId:             item.UserId,
			SuccessOrderCount:  int64(item.SuccessOrderCount),
			SuccessTotalAmount: item.SuccessTotalAmount,
			TodaySuccessAmount: item.TodaySuccessAmount,
			TodaySuccessCount:  int64(item.TodaySuccessCount),
			FirstSuccessTime:   item.FirstSuccessTime,
			LastSuccessTime:    item.LastSuccessTime,
			CreateTime:         item.CreateTime,
			UpdateTime:         item.UpdateTime,
		}
	}

	return &types.ListUserRechargeStatsResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
