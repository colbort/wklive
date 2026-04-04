package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateWithdrawOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWithdrawOrderLogic {
	return &CreateWithdrawOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提现
func (l *CreateWithdrawOrderLogic) CreateWithdrawOrder(in *payment.CreateWithdrawOrderReq) (*payment.CreateWithdrawOrderResp, error) {
	// todo: add your logic here and delete this line

	return &payment.CreateWithdrawOrderResp{}, nil
}
