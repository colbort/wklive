package logic

import (
	"context"
	"errors"

	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessPositionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessPositionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessPositionsLogic {
	return &ProcessPositionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 仓位处理（标记价格刷新/强平扫描/普通平仓）
func (l *ProcessPositionsLogic) ProcessPositions(in *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
	return runTradeTaskWithLock(l.ctx, l.svcCtx, "process_positions", func() (*trade.TradeTaskResp, error) {
		if err := l.refreshMarkPrices(in); err != nil {
			return nil, err
		}
		if err := l.forceLiquidation(in); err != nil {
			return nil, err
		}
		if err := l.closePositions(in); err != nil {
			return nil, err
		}
		return okTradeTaskResp(), nil
	})
}

func (l *ProcessPositionsLogic) refreshMarkPrices(in *trade.TradeTaskReq) error {
	cursor := int64(0)
	for {
		positions, _, err := l.svcCtx.ContractPositionModel.FindPage(l.ctx, models.ContractPositionPageFilter{TenantId: in.GetTenantId()}, cursor, 100)
		if err != nil {
			return err
		}
		if len(positions) == 0 {
			return nil
		}
		for _, position := range positions {
			cursor = position.Id
			if position.Qty <= 0 {
				continue
			}
			if err := createTradeTaskEvent(l.ctx, l.svcCtx, position.TenantId, "MARK_PRICE_REFRESH_REQUIRED", "position", position.Id, position.UserId, position.SymbolId, position.MarketType, "mark price refresh task"); err != nil {
				return err
			}
		}
		if len(positions) < 100 {
			return nil
		}
	}
}

func (l *ProcessPositionsLogic) forceLiquidation(in *trade.TradeTaskReq) error {
	cursor := int64(0)
	for {
		positions, _, err := l.svcCtx.ContractPositionModel.FindPage(l.ctx, models.ContractPositionPageFilter{TenantId: in.GetTenantId()}, cursor, 100)
		if err != nil {
			return err
		}
		if len(positions) == 0 {
			return nil
		}
		for _, position := range positions {
			cursor = position.Id
			if position.Qty <= 0 || position.MarkPrice <= 0 || position.LiquidationPrice <= 0 {
				continue
			}
			needLiquidation := (position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_LONG) && position.MarkPrice <= position.LiquidationPrice) ||
				(position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_SHORT) && position.MarkPrice >= position.LiquidationPrice)
			if !needLiquidation {
				continue
			}
			if err := createTradeTaskEvent(l.ctx, l.svcCtx, position.TenantId, "FORCE_LIQUIDATION_REQUIRED", "position", position.Id, position.UserId, position.SymbolId, position.MarketType, "force liquidation task"); err != nil {
				return err
			}
		}
		if len(positions) < 100 {
			return nil
		}
	}
}

func (l *ProcessPositionsLogic) closePositions(in *trade.TradeTaskReq) error {
	now := utils.NowMillis()
	cursor := int64(0)
	for {
		positions, _, err := l.svcCtx.ContractPositionModel.FindPage(l.ctx, models.ContractPositionPageFilter{TenantId: in.GetTenantId()}, cursor, 100)
		if err != nil {
			return err
		}
		if len(positions) == 0 {
			return nil
		}
		for _, position := range positions {
			cursor = position.Id
			if position.Qty <= 0 {
				continue
			}
			symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, position.SymbolId)
			if errors.Is(err, models.ErrNotFound) {
				continue
			}
			if err != nil {
				return err
			}
			if symbol.Status != int64(trade.SymbolStatus_SYMBOL_STATUS_DISABLED) && (symbol.CloseTime == 0 || symbol.CloseTime > now) {
				continue
			}
			if err := createTradeTaskEvent(l.ctx, l.svcCtx, position.TenantId, "CLOSE_POSITION_REQUIRED", "position", position.Id, position.UserId, position.SymbolId, position.MarketType, "close position task"); err != nil {
				return err
			}
		}
		if len(positions) < 100 {
			return nil
		}
	}
}
