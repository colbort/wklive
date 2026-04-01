package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyRechargeOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMyRechargeOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyRechargeOrderStatusLogic {
	return &QueryMyRechargeOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 轮询订单状态
func (l *QueryMyRechargeOrderStatusLogic) QueryMyRechargeOrderStatus(in *payment.QueryMyRechargeOrderStatusReq) (*payment.QueryMyRechargeOrderStatusResp, error) {
	// todo: add your logic here and delete this line

	return &payment.QueryMyRechargeOrderStatusResp{}, nil
}
