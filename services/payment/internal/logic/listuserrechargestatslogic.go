package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
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

	lastID := int64(0)
	if len(stats) > 0 {
		lastID = stats[len(stats)-1].Id
	}

	data := make([]*payment.UserRechargeStat, 0, len(stats))
	for _, s := range stats {
		data = append(data, toUserRechargeStatProto(s))
	}

	return &payment.ListUserRechargeStatsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(stats), total, lastID),
		Data: data,
	}, nil
}
