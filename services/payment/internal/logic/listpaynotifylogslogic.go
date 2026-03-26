package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPayNotifyLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPayNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayNotifyLogsLogic {
	return &ListPayNotifyLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 回调日志列表
func (l *ListPayNotifyLogsLogic) ListPayNotifyLogs(in *payment.ListPayNotifyLogsReq) (*payment.ListPayNotifyLogsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListPayNotifyLogsResp{}, nil
}
