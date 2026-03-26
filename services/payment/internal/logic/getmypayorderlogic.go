package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyPayOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyPayOrderLogic {
	return &GetMyPayOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的订单详情
func (l *GetMyPayOrderLogic) GetMyPayOrder(in *payment.GetMyPayOrderReq) (*payment.GetMyPayOrderResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetMyPayOrderResp{}, nil
}
