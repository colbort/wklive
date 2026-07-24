package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PauseSymbolLiquidityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPauseSymbolLiquidityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PauseSymbolLiquidityLogic {
	return &PauseSymbolLiquidityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PauseSymbolLiquidityLogic) PauseSymbolLiquidity(in *liquidity.SymbolActionReq) (*liquidity.CommonResp, error) {
	if err := changeSymbolStatus(l.ctx, l.svcCtx, in, liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_PAUSED); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
