package adminlogic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &liquidity.CommonResp{}, nil
}
