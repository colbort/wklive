package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRechargeOrderLogic {
	return &CreateRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建充值订单
func (l *CreateRechargeOrderLogic) CreateRechargeOrder(in *payment.CreateRechargeOrderReq) (*payment.CreateRechargeOrderResp, error) {
	// todo: add your logic here and delete this line

	return &payment.CreateRechargeOrderResp{}, nil
}
