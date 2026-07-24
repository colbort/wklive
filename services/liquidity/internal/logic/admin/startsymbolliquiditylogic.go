package adminlogic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &liquidity.CommonResp{}, nil
}
