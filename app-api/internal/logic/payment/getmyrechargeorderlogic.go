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

type GetMyRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyRechargeOrderLogic {
	return &GetMyRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyRechargeOrderLogic) GetMyRechargeOrder(req *types.GetMyRechargeOrderReq) (resp *types.GetMyRechargeOrderResp, err error) {
	return logicutil.Proxy[types.GetMyRechargeOrderResp](l.ctx, req, l.svcCtx.PaymentCli.GetMyRechargeOrder)
}
