package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualMarkRechargeOrderSuccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualMarkRechargeOrderSuccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualMarkRechargeOrderSuccessLogic {
	return &ManualMarkRechargeOrderSuccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 人工标记充值订单支付成功
func (l *ManualMarkRechargeOrderSuccessLogic) ManualMarkRechargeOrderSuccess(in *payment.ManualMarkRechargeOrderSuccessReq) (*payment.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &payment.AdminCommonResp{}, nil
}
