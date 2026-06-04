// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

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
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.PaymentCli.ManualMarkRechargeOrderSuccess)
}
