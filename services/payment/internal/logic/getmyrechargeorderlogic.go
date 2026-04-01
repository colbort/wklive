package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyRechargeOrderLogic {
	return &GetMyRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的订单详情
func (l *GetMyRechargeOrderLogic) GetMyRechargeOrder(in *payment.GetMyRechargeOrderReq) (*payment.GetMyRechargeOrderResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetMyRechargeOrderResp{}, nil
}
