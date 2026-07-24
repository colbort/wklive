// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExternalOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExternalOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExternalOrderListLogic {
	return &ExternalOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExternalOrderListLogic) ExternalOrderList(req *types.OrderQuery) (resp *types.OrderListResp, err error) {
	return orderList(l.ctx, l.svcCtx, req, true)
}
