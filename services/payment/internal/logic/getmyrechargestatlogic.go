package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyRechargeStatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyRechargeStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyRechargeStatLogic {
	return &GetMyRechargeStatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 当前用户累计充值统计
func (l *GetMyRechargeStatLogic) GetMyRechargeStat(in *payment.GetMyRechargeStatReq) (*payment.GetMyRechargeStatResp, error) {
	stat, err := l.svcCtx.UserRechargeStatModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if stat == nil {
		// Return default stat if not found
		return &payment.GetMyRechargeStatResp{
			Base: helper.OkResp(),
			Stat: &payment.UserRechargeStat{
				TenantId: in.TenantId,
				UserId:   in.UserId,
			},
		}, nil
	}

	return &payment.GetMyRechargeStatResp{
		Base: helper.OkResp(),
		Stat: &payment.UserRechargeStat{
			Id:                 stat.Id,
			TenantId:           stat.TenantId,
			UserId:             stat.UserId,
			SuccessOrderCount:  stat.SuccessOrderCount,
			SuccessTotalAmount: stat.SuccessTotalAmount,
			TodaySuccessAmount: stat.TodaySuccessAmount,
			TodaySuccessCount:  stat.TodaySuccessCount,
			FirstSuccessTime:   stat.FirstSuccessTime.Int64,
			LastSuccessTime:    stat.LastSuccessTime.Int64,
			CreateTimes:        stat.CreateTimes,
			UpdateTimes:        stat.UpdateTimes,
		},
	}, nil
}
