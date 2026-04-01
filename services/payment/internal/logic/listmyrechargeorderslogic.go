package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyRechargeOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyRechargeOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyRechargeOrdersLogic {
	return &ListMyRechargeOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的充值订单列表
func (l *ListMyRechargeOrdersLogic) ListMyRechargeOrders(in *payment.ListMyRechargeOrdersReq) (*payment.ListMyRechargeOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListMyRechargeOrdersResp{}, nil
}
