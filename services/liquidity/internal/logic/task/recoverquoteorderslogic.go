package tasklogic

import (
	"context"
	"wklive/services/liquidity/internal/logic/helpers"

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
	if err := helpers.ValidateTask(in); err != nil {
		return nil, err
	}
	return processInternalQuotes(l.ctx, l.svcCtx, in, true)
}
