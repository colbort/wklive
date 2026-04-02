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

type GetUserRechargeStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserRechargeStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRechargeStatLogic {
	return &GetUserRechargeStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserRechargeStatLogic) GetUserRechargeStat(req *types.GetUserRechargeStatReq) (resp *types.GetUserRechargeStatResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetUserRechargeStat(l.ctx, &payment.GetUserRechargeStatReq{
		TenantId: req.TenantId,
		UserId:   req.UserId,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetUserRechargeStatResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.UserRechargeStat{
			Id:                 result.Data.Id,
			TenantId:           result.Data.TenantId,
			UserId:             result.Data.UserId,
			SuccessOrderCount:  int64(result.Data.SuccessOrderCount),
			SuccessTotalAmount: result.Data.SuccessTotalAmount,
			TodaySuccessAmount: result.Data.TodaySuccessAmount,
			TodaySuccessCount:  int64(result.Data.TodaySuccessCount),
			FirstSuccessTime:   result.Data.FirstSuccessTime,
			LastSuccessTime:    result.Data.LastSuccessTime,
			CreateTime:         result.Data.CreateTime,
			UpdateTime:         result.Data.UpdateTime,
		},
	}, nil
}
