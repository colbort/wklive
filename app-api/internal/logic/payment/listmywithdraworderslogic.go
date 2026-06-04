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

type ListMyWithdrawOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyWithdrawOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyWithdrawOrdersLogic {
	return &ListMyWithdrawOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyWithdrawOrdersLogic) ListMyWithdrawOrders(req *types.ListMyWithdrawOrdersReq) (resp *types.ListMyWithdrawOrdersResp, err error) {
	return logicutil.Proxy[types.ListMyWithdrawOrdersResp](l.ctx, req, l.svcCtx.PaymentCli.ListMyWithdrawOrders)
}
