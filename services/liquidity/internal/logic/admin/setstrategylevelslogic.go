package adminlogic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetStrategyLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetStrategyLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetStrategyLevelsLogic {
	return &SetStrategyLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetStrategyLevelsLogic) SetStrategyLevels(in *liquidity.SetStrategyLevelsReq) (*liquidity.CommonResp, error) {
	if in.ConfigId <= 0 {
		return nil, fmt.Errorf("config_id are required")
	}
	if len(in.Levels) == 0 {
		return nil, fmt.Errorf("at least one strategy level is required")
	}
	config, err := l.svcCtx.SymbolConfigModel.FindOne(l.ctx, in.ConfigId)
	if err != nil {
		return nil, err
	}
	if config.Version != in.ConfigVersion {
		return nil, fmt.Errorf("symbol config version conflict")
	}
	if config.Status == int64(liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING) {
		return nil, fmt.Errorf("pause liquidity before updating strategy levels")
	}
	seen := make(map[int32]struct{}, len(in.Levels))
	parsed := make([]models.LiquidityStrategyLevelInput, 0, len(in.Levels))
	for _, level := range in.Levels {
		if level.LevelNo <= 0 {
			return nil, fmt.Errorf("level_no must be positive")
		}
		if _, ok := seen[level.LevelNo]; ok {
			return nil, fmt.Errorf("duplicate level_no %d", level.LevelNo)
		}
		seen[level.LevelNo] = struct{}{}
		bidSpread, err := parseNumber("bid_spread_bps", level.BidSpreadBps)
		if err != nil {
			return nil, err
		}
		askSpread, err := parseNumber("ask_spread_bps", level.AskSpreadBps)
		if err != nil {
			return nil, err
		}
		bidQty, err := parseNumber("bid_qty", level.BidQty)
		if err != nil || bidQty <= 0 {
			return nil, fmt.Errorf("bid_qty must be positive")
		}
		askQty, err := parseNumber("ask_qty", level.AskQty)
		if err != nil || askQty <= 0 {
			return nil, fmt.Errorf("ask_qty must be positive")
		}
		parsed = append(parsed, models.LiquidityStrategyLevelInput{
			LevelNo:      int64(level.LevelNo),
			BidSpreadBps: bidSpread,
			AskSpreadBps: askSpread,
			BidQty:       bidQty,
			AskQty:       askQty,
			Enabled:      int64(level.Enabled),
		})
	}
	now := time.Now().UnixMilli()
	if err := l.svcCtx.StrategyLevelModel.Replace(l.ctx, in.ConfigId, parsed, now); err != nil {
		return nil, err
	}
	config.Version++
	config.UpdateTimes = now
	if err := l.svcCtx.SymbolConfigModel.Update(l.ctx, config); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
