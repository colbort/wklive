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

type CloseRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCloseRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseRechargeOrderLogic {
	return &CloseRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CloseRechargeOrderLogic) CloseRechargeOrder(req *types.CloseRechargeOrderReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.CloseRechargeOrder(l.ctx, &payment.CloseRechargeOrderReq{
		TenantId: req.TenantId,
		OrderNo:  req.OrderNo,
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
