// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAllOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelAllOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAllOrdersLogic {
	return &CancelAllOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelAllOrdersLogic) CancelAllOrders(req *types.CancelAllOrdersReq) (resp *types.CancelAllOrdersResp, err error) {
	return logicutil.Proxy[types.CancelAllOrdersResp](l.ctx, req, l.svcCtx.TradeCli.CancelAllOrders)
}
