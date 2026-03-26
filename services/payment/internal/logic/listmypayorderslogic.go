package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyPayOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyPayOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyPayOrdersLogic {
	return &ListMyPayOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的充值订单列表
func (l *ListMyPayOrdersLogic) ListMyPayOrders(in *payment.ListMyPayOrdersReq) (*payment.ListMyPayOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListMyPayOrdersResp{}, nil
}
