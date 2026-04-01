package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRechargeOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRechargeOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRechargeOrdersLogic {
	return &ListRechargeOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 充值订单列表
func (l *ListRechargeOrdersLogic) ListRechargeOrders(in *payment.ListRechargeOrdersReq) (*payment.ListRechargeOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListRechargeOrdersResp{}, nil
}
