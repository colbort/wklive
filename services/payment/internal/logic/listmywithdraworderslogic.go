package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyWithdrawOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyWithdrawOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyWithdrawOrdersLogic {
	return &ListMyWithdrawOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现订单列表
func (l *ListMyWithdrawOrdersLogic) ListMyWithdrawOrders(in *payment.ListMyWithdrawOrdersReq) (*payment.ListMyWithdrawOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListMyWithdrawOrdersResp{}, nil
}
