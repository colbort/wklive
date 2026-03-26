package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRechargeStatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserRechargeStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRechargeStatLogic {
	return &GetUserRechargeStatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户充值统计
func (l *GetUserRechargeStatLogic) GetUserRechargeStat(in *payment.GetUserRechargeStatReq) (*payment.GetUserRechargeStatResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetUserRechargeStatResp{}, nil
}
