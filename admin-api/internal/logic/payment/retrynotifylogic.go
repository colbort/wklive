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

type RetryNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryNotifyLogic {
	return &RetryNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryNotifyLogic) RetryNotify(req *types.RetryNotifyReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.RetryNotify(l.ctx, &payment.RetryNotifyReq{
		TenantId: req.TenantId,
		OrderNo:  req.OrderNo,
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
