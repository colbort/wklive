package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRechargeNotifyLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRechargeNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRechargeNotifyLogLogic {
	return &GetRechargeNotifyLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 充值回调日志
func (l *GetRechargeNotifyLogLogic) GetRechargeNotifyLog(in *payment.GetRechargeNotifyLogReq) (*payment.GetRechargeNotifyLogResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetRechargeNotifyLogResp{}, nil
}
