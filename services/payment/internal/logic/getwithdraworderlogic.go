package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWithdrawOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawOrderLogic {
	return &GetWithdrawOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现订单详情
func (l *GetWithdrawOrderLogic) GetWithdrawOrder(in *payment.GetWithdrawOrderReq) (*payment.GetWithdrawOrderResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetWithdrawOrderResp{}, nil
}
