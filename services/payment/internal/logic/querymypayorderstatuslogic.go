package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyPayOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMyPayOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyPayOrderStatusLogic {
	return &QueryMyPayOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 轮询订单状态
func (l *QueryMyPayOrderStatusLogic) QueryMyPayOrderStatus(in *payment.QueryMyPayOrderStatusReq) (*payment.QueryMyPayOrderStatusResp, error) {
	// todo: add your logic here and delete this line

	return &payment.QueryMyPayOrderStatusResp{}, nil
}
