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

type ManualMarkRechargeOrderSuccessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualMarkRechargeOrderSuccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualMarkRechargeOrderSuccessLogic {
	return &ManualMarkRechargeOrderSuccessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ManualMarkRechargeOrderSuccessLogic) ManualMarkRechargeOrderSuccess(req *types.ManualMarkRechargeOrderSuccessReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.ManualMarkRechargeOrderSuccess(l.ctx, &payment.ManualMarkRechargeOrderSuccessReq{
		TenantId:     req.TenantId,
		OrderNo:      req.OrderNo,
		ThirdTradeNo: req.ThirdTradeNo,
		PayAmount:    req.PayAmount,
		Remark:       req.Remark,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}
	return resp, nil
}
