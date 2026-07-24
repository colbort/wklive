package adminlogic

import (
	"context"
	"fmt"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolConfigDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolConfigDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolConfigDetailLogic {
	return &GetSymbolConfigDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSymbolConfigDetailLogic) GetSymbolConfigDetail(in *liquidity.GetSymbolConfigDetailReq) (*liquidity.GetSymbolConfigDetailResp, error) {
	if in.Id <= 0 && in.SymbolId <= 0 {
		return nil, fmt.Errorf("id or symbol_id is required")
	}
	row, err := l.svcCtx.SymbolConfigModel.FindByIDOrSymbol(l.ctx, in.Id, in.SymbolId)
	if err != nil {
		return nil, err
	}
	levels, err := l.svcCtx.StrategyLevelModel.FindList(l.ctx, row.Id, false)
	if err != nil {
		return nil, err
	}
	levelData := make([]*liquidity.LiquidityStrategyLevel, 0, len(levels))
	for _, level := range levels {
		levelData = append(levelData, helpers.StrategyLevelToProto(level))
	}
	return &liquidity.GetSymbolConfigDetailResp{Base: helper.OkResp(), Data: helpers.SymbolConfigToProto(row), Levels: levelData}, nil
}
