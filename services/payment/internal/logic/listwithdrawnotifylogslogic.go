package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWithdrawNotifyLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWithdrawNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawNotifyLogsLogic {
	return &ListWithdrawNotifyLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提现回调日志列表
func (l *ListWithdrawNotifyLogsLogic) ListWithdrawNotifyLogs(in *payment.ListWithdrawNotifyLogsReq) (*payment.ListWithdrawNotifyLogsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListWithdrawNotifyLogsResp{}, nil
}
