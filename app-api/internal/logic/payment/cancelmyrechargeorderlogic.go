// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"wklive/app-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelMyRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelMyRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelMyRechargeOrderLogic {
	return &CancelMyRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelMyRechargeOrderLogic) CancelMyRechargeOrder(req *types.CancelMyRechargeOrderReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.PaymentCli.CancelMyRechargeOrder)
}
