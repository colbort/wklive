package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelMyRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelMyRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelMyRechargeOrderLogic {
	return &CancelMyRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取消未支付订单
func (l *CancelMyRechargeOrderLogic) CancelMyRechargeOrder(in *payment.CancelMyRechargeOrderReq) (*payment.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &payment.AppCommonResp{}, nil
}
