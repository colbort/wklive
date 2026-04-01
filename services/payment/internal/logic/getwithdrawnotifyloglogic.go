package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWithdrawNotifyLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWithdrawNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawNotifyLogLogic {
	return &GetWithdrawNotifyLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现回调日志详情
func (l *GetWithdrawNotifyLogLogic) GetWithdrawNotifyLog(in *payment.GetWithdrawNotifyLogReq) (*payment.GetWithdrawNotifyLogResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetWithdrawNotifyLogResp{}, nil
}
