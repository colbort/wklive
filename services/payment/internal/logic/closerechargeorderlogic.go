package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CloseRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseRechargeOrderLogic {
	return &CloseRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 关闭充值订单
func (l *CloseRechargeOrderLogic) CloseRechargeOrder(in *payment.CloseRechargeOrderReq) (*payment.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &payment.AdminCommonResp{}, nil
}
