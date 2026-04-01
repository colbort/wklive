package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuditWithdrawOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuditWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditWithdrawOrderLogic {
	return &AuditWithdrawOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 审核提现订单
func (l *AuditWithdrawOrderLogic) AuditWithdrawOrder(in *payment.AuditWithdrawOrderReq) (*payment.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &payment.AdminCommonResp{}, nil
}
