package tasklogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecoverQuoteOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecoverQuoteOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecoverQuoteOrdersLogic {
	return &RecoverQuoteOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RecoverQuoteOrdersLogic) RecoverQuoteOrders(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.LiquidityTaskResp{}, nil
}
