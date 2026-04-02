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

type GetMyRechargeStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyRechargeStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyRechargeStatLogic {
	return &GetMyRechargeStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyRechargeStatLogic) GetMyRechargeStat(req *types.GetMyRechargeStatReq) (resp *types.GetMyRechargeStatResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.PaymentCli.GetMyRechargeStat(l.ctx, &payment.GetMyRechargeStatReq{
		TenantId: req.TenantId,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetMyRechargeStatResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.UserRechargeStat{
			Id:                 result.Stat.Id,
			TenantId:           result.Stat.TenantId,
			UserId:             result.Stat.UserId,
			SuccessOrderCount:  result.Stat.SuccessOrderCount,
			SuccessTotalAmount: result.Stat.SuccessTotalAmount,
			TodaySuccessAmount: result.Stat.TodaySuccessAmount,
			TodaySuccessCount:  result.Stat.TodaySuccessCount,
			FirstSuccessTime:   result.Stat.FirstSuccessTime,
			LastSuccessTime:    result.Stat.LastSuccessTime,
			CreateTime:         result.Stat.CreateTime,
			UpdateTime:         result.Stat.UpdateTime,
		},
	}

	return
}
