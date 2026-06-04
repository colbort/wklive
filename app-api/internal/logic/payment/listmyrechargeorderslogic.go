// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/logicutil"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyRechargeOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyRechargeOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyRechargeOrdersLogic {
	return &ListMyRechargeOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyRechargeOrdersLogic) ListMyRechargeOrders(req *types.ListMyRechargeOrdersReq) (resp *types.ListMyRechargeOrdersResp, err error) {
	return logicutil.Proxy[types.ListMyRechargeOrdersResp](l.ctx, req, l.svcCtx.PaymentCli.ListMyRechargeOrders)
}
