package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelMyPayOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelMyPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelMyPayOrderLogic {
	return &CancelMyPayOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取消未支付订单
func (l *CancelMyPayOrderLogic) CancelMyPayOrder(in *payment.CancelMyPayOrderReq) (*payment.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &payment.AppCommonResp{}, nil
}
