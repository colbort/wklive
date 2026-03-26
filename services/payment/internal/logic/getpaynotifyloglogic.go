package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPayNotifyLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPayNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayNotifyLogLogic {
	return &GetPayNotifyLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 回调日志
func (l *GetPayNotifyLogLogic) GetPayNotifyLog(in *payment.GetPayNotifyLogReq) (*payment.GetPayNotifyLogResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetPayNotifyLogResp{}, nil
}
