// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRechargeOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRechargeOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRechargeOrdersLogic {
	return &ListRechargeOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRechargeOrdersLogic) ListRechargeOrders(req *types.ListRechargeOrdersReq) (resp *types.ListRechargeOrdersResp, err error) {
	return logicutil.Proxy[types.ListRechargeOrdersResp](l.ctx, req, l.svcCtx.PaymentCli.ListRechargeOrders)
}
