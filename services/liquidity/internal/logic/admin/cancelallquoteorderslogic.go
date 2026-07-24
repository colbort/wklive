package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAllQuoteOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelAllQuoteOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAllQuoteOrdersLogic {
	return &CancelAllQuoteOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelAllQuoteOrdersLogic) CancelAllQuoteOrders(in *liquidity.SymbolActionReq) (*liquidity.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.CommonResp{}, nil
}
