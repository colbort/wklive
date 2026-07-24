package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartSymbolLiquidityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartSymbolLiquidityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartSymbolLiquidityLogic {
	return &StartSymbolLiquidityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StartSymbolLiquidityLogic) StartSymbolLiquidity(in *liquidity.SymbolActionReq) (*liquidity.CommonResp, error) {
	if err := changeSymbolStatus(l.ctx, l.svcCtx, in, liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
