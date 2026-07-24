package liquiditylogic

import (
	"context"
	"fmt"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActiveSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActiveSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActiveSymbolConfigLogic {
	return &GetActiveSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetActiveSymbolConfigLogic) GetActiveSymbolConfig(in *liquidity.GetActiveSymbolConfigReq) (*liquidity.GetSymbolConfigDetailResp, error) {
	if in.SymbolId <= 0 {
		return nil, fmt.Errorf("symbol_id are required")
	}
	row, err := l.svcCtx.SymbolConfigModel.FindActiveBySymbol(l.ctx, in.SymbolId)
	if err != nil {
		return nil, err
	}
	levels, err := l.svcCtx.StrategyLevelModel.FindList(l.ctx, row.Id, true)
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityStrategyLevel, 0, len(levels))
	for _, level := range levels {
		data = append(data, helpers.StrategyLevelToProto(level))
	}
	return &liquidity.GetSymbolConfigDetailResp{Base: helper.OkResp(), Data: helpers.SymbolConfigToProto(row), Levels: data}, nil
}
