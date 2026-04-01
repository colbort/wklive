package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyWithdrawOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyWithdrawOrderLogic {
	return &GetMyWithdrawOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现订单详情
func (l *GetMyWithdrawOrderLogic) GetMyWithdrawOrder(in *payment.GetMyWithdrawOrderReq) (*payment.GetMyWithdrawOrderResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetMyWithdrawOrderResp{}, nil
}
