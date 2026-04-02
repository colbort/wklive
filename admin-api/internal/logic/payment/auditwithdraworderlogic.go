// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuditWithdrawOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuditWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditWithdrawOrderLogic {
	return &AuditWithdrawOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuditWithdrawOrderLogic) AuditWithdrawOrder(req *types.AuditWithdrawOrderReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.AuditWithdrawOrder(l.ctx, &payment.AuditWithdrawOrderReq{
		TenantId: req.TenantId,
		OrderNo:  req.OrderNo,
		Approve:  req.Approve,
		Remark:   req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}
