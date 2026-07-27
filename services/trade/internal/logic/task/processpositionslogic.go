package tasklogic

import (
	"context"
	"errors"
	"wklive/services/trade/internal/logic/helpers"

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
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_positions", func() (*trade.TradeTaskResp, error) {
		if err := l.refreshMarkPrices(in); err != nil {
			return nil, err
		}
		if err := l.forceLiquidation(in); err != nil {
			return nil, err
		}
		return helpers.OkTaskResp(), nil
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
			if !position.Qty.IsPositive() {
				continue
			}
			contract, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, position.TenantId, position.SymbolId)
			if errors.Is(err, models.ErrNotFound) {
				continue
			}
			if err != nil {
				return err
			}
			quote, err := NewProcessSecondsSettlementsLogic(l.ctx, l.svcCtx).getValidQuoteKind("MARK_PRICE", contract.MarkPriceSource, position.SymbolId, 30_000)
			if err != nil {
				l.Errorf("skip stale mark price, positionId=%d err=%v", position.Id, err)
				continue
			}
			position.MarkPrice = helpers.MustParseFloat(quote.LastPrice)
			position.MarkSnapshotId = quote.SnapshotID
			tier, err := NewProcessContractPositionFillsLogic(l.ctx, l.svcCtx).riskTierForPosition(l.ctx, position, contract)
			if err != nil {
				return err
			}
			recalculatePositionRisk(position, contract, tier)
			position.Version++
			position.UpdateTimes = utils.NowMillis()
			if err := l.svcCtx.ContractPositionModel.Update(l.ctx, position); err != nil {
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
			if !position.Qty.IsPositive() || !position.MarkPrice.IsPositive() || !position.LiquidationPrice.IsPositive() {
				continue
			}
			needLiquidation := (position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_LONG) && position.MarkPrice.LessThanOrEqual(position.LiquidationPrice)) ||
				(position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_SHORT) && position.MarkPrice.GreaterThanOrEqual(position.LiquidationPrice))
			if !needLiquidation {
				continue
			}
			if err := NewProcessLiquidationsLogic(l.ctx, l.svcCtx).ProcessPosition(position.Id); err != nil {
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
			if !position.Qty.IsPositive() {
				continue
			}
			symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, position.SymbolId)
			if errors.Is(err, models.ErrNotFound) {
				continue
			}
			if err != nil {
				return err
			}
			if symbol.Status != int64(trade.SymbolStatus_SYMBOL_STATUS_DISABLED) && (symbol.TradingEndTime == 0 || symbol.TradingEndTime > now) {
				continue
			}
			if err := helpers.CreateTaskEvent(l.ctx, l.svcCtx, position.TenantId, "CLOSE_POSITION_REQUIRED", "position", position.Id, position.UserId, position.SymbolId, symbol.ProductType, "close position task"); err != nil {
				return err
			}
		}
		if len(positions) < 100 {
			return nil
		}
	}
}
