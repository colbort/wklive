// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAllQuoteOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelAllQuoteOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAllQuoteOrdersLogic {
	return &CancelAllQuoteOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelAllQuoteOrdersLogic) CancelAllQuoteOrders(req *types.CancelAllQuoteOrdersReq) (resp *types.RespBase, err error) {
	return cancelAllQuoteOrders(l.ctx, l.svcCtx, req)
}
