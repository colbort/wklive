package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualMarkPayOrderSuccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualMarkPayOrderSuccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualMarkPayOrderSuccessLogic {
	return &ManualMarkPayOrderSuccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 人工标记订单支付成功
func (l *ManualMarkPayOrderSuccessLogic) ManualMarkPayOrderSuccess(in *payment.ManualMarkPayOrderSuccessReq) (*payment.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &payment.AdminCommonResp{}, nil
}
