package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

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
	// todo: add your logic here and delete this line

	return &payment.ListUserRechargeStatsResp{}, nil
}
