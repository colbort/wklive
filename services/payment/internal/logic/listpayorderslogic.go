package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPayOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPayOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayOrdersLogic {
	return &ListPayOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订单列表
func (l *ListPayOrdersLogic) ListPayOrders(in *payment.ListPayOrdersReq) (*payment.ListPayOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListPayOrdersResp{}, nil
}
