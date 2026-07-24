package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type StopSymbolLiquidityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStopSymbolLiquidityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StopSymbolLiquidityLogic {
	return &StopSymbolLiquidityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StopSymbolLiquidityLogic) StopSymbolLiquidity(in *liquidity.SymbolActionReq) (*liquidity.CommonResp, error) {
	if err := changeSymbolStatus(l.ctx, l.svcCtx, in, liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_DISABLED); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
