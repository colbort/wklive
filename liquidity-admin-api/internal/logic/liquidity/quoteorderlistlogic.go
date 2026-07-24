// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QuoteOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQuoteOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuoteOrderListLogic {
	return &QuoteOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QuoteOrderListLogic) QuoteOrderList(req *types.OrderQuery) (resp *types.OrderListResp, err error) {
	return orderList(l.ctx, l.svcCtx, req, false)
}
