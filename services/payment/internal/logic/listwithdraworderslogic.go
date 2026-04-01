package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWithdrawOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWithdrawOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawOrdersLogic {
	return &ListWithdrawOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现订单列表
func (l *ListWithdrawOrdersLogic) ListWithdrawOrders(in *payment.ListWithdrawOrdersReq) (*payment.ListWithdrawOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListWithdrawOrdersResp{}, nil
}
