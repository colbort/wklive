package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRechargeNotifyLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRechargeNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRechargeNotifyLogsLogic {
	return &ListRechargeNotifyLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 充值回调日志列表
func (l *ListRechargeNotifyLogsLogic) ListRechargeNotifyLogs(in *payment.ListRechargeNotifyLogsReq) (*payment.ListRechargeNotifyLogsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListRechargeNotifyLogsResp{}, nil
}
