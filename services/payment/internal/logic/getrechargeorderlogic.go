package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRechargeOrderLogic {
	return &GetRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取充值订单
func (l *GetRechargeOrderLogic) GetRechargeOrder(in *payment.GetRechargeOrderReq) (*payment.GetRechargeOrderResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetRechargeOrderResp{}, nil
}
