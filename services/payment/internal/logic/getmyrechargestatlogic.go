package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

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
	// todo: add your logic here and delete this line

	return &payment.GetMyRechargeStatResp{}, nil
}
