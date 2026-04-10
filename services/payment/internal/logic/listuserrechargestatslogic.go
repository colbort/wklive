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

type ListUserRechargeStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserRechargeStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserRechargeStatsLogic {
	return &ListUserRechargeStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户充值统计列表
func (l *ListUserRechargeStatsLogic) ListUserRechargeStats(in *payment.ListUserRechargeStatsReq) (*payment.ListUserRechargeStatsResp, error) {
	stats, total, err := l.svcCtx.UserRechargeStatModel.FindPage(
		l.ctx,
		in.TenantId,
		in.UserId,
		in.SuccessTotalAmountMin,
		in.SuccessTotalAmountMax,
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(stats)) == in.Page.Limit {
		lastItem := stats[len(stats)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(stats)) == in.Page.Limit

	data := make([]*payment.UserRechargeStat, 0, len(stats))
	for _, s := range stats {
		data = append(data, &payment.UserRechargeStat{
			Id:                 s.Id,
			TenantId:           s.TenantId,
			UserId:             s.UserId,
			SuccessOrderCount:  s.SuccessOrderCount,
			SuccessTotalAmount: s.SuccessTotalAmount,
			TodaySuccessAmount: s.TodaySuccessAmount,
			TodaySuccessCount:  s.TodaySuccessCount,
			FirstSuccessTime:   s.FirstSuccessTime.Int64,
			LastSuccessTime:    s.LastSuccessTime.Int64,
			CreateTimes:        s.CreateTimes,
			UpdateTimes:        s.UpdateTimes,
		})
	}

	return &payment.ListUserRechargeStatsResp{
		Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor),
		Data: data,
	}, nil
}
