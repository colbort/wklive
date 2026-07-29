package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const markQuoteRetryBackoffMs int64 = 30_000

type markQuoteRetryGate struct {
	mu         sync.Mutex
	retryAfter map[string]int64
}

func newMarkQuoteRetryGate() *markQuoteRetryGate {
	return &markQuoteRetryGate{retryAfter: make(map[string]int64)}
}

func (g *markQuoteRetryGate) allow(key string, now int64) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	retryAt, found := g.retryAfter[key]
	if !found || retryAt <= now {
		delete(g.retryAfter, key)
		return true
	}
	return false
}

func (g *markQuoteRetryGate) fail(key string, now int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.retryAfter[key] = now + markQuoteRetryBackoffMs
}

func (g *markQuoteRetryGate) success(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.retryAfter, key)
}

var processMarkQuoteRetryGate = newMarkQuoteRetryGate()

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
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_positions", func(taskCtx context.Context) (*trade.TradeTaskResp, error) {
		l.ctx = taskCtx
		if err := l.refreshMarkPrices(in); err != nil {
			return nil, err
		}
		var result error
		if err := l.refreshCrossMarginSnapshots(in.GetTenantId()); err != nil {
			result = errors.Join(result, fmt.Errorf("cross margin risk projection: %w", err))
		}
		if err := NewProcessCrossMarginLiquidationsLogic(l.ctx, l.svcCtx).ProcessRiskSnapshots(in.GetTenantId()); err != nil {
			result = errors.Join(result, fmt.Errorf("cross margin account liquidation: %w", err))
		}
		if err := l.forceLiquidation(in); err != nil {
			result = errors.Join(result, err)
		}
		if result != nil {
			return nil, result
		}
		return helpers.OkTaskResp(), nil
	})
}

func (l *ProcessPositionsLogic) refreshMarkPrices(in *trade.TradeTaskReq) error {
	type quoteLookup struct {
		quote *marketQuoteSnapshot
		ok    bool
	}
	// A symbol can have several user/side positions. Resolve its authoritative
	// MARK once per scan and share the result instead of issuing one archive RPC
	// per position.
	quotes := make(map[string]quoteLookup)
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
			if !position.Qty.IsPositive() || position.Status != int64(trade.PositionStatus_POSITION_STATUS_NORMAL) {
				continue
			}
			contract, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, position.TenantId, position.SymbolId)
			if errors.Is(err, models.ErrNotFound) {
				continue
			}
			if err != nil {
				return err
			}

			quoteKey := fmt.Sprintf("%d:%d:%s", position.TenantId, position.SymbolId, contract.MarkPriceSource)
			lookup, found := quotes[quoteKey]
			if !found {
				now := utils.NowMillis()
				if !processMarkQuoteRetryGate.allow(quoteKey, now) {
					quotes[quoteKey] = quoteLookup{}
					continue
				}
				quote, quoteErr := NewProcessSecondsSettlementsLogic(l.ctx, l.svcCtx).
					getValidQuoteKind("MARK_PRICE", contract.MarkPriceSource, position.SymbolId, 30_000)
				if quoteErr != nil {
					processMarkQuoteRetryGate.fail(quoteKey, now)
					l.Errorf(
						"skip stale mark price, tenantId=%d symbolId=%d source=%s err=%v",
						position.TenantId, position.SymbolId, contract.MarkPriceSource, quoteErr,
					)
					quotes[quoteKey] = quoteLookup{}
					continue
				}
				processMarkQuoteRetryGate.success(quoteKey)
				lookup = quoteLookup{quote: quote, ok: true}
				quotes[quoteKey] = lookup
			}
			if !lookup.ok {
				continue
			}
			quote := lookup.quote
			beforeRisk := *position
			position.MarkPrice = helpers.MustParseFloat(quote.LastPrice)
			position.MarkSnapshotId = quote.SnapshotID
			tier, err := NewProcessContractPositionFillsLogic(l.ctx, l.svcCtx).riskTierForPosition(l.ctx, position, contract)
			if err != nil {
				return err
			}
			expectedVersion := position.Version
			recalculatePositionRisk(position, contract, tier)
			if markRiskProjectionEqual(&beforeRisk, position) {
				continue
			}
			updated, err := l.svcCtx.ContractPositionModel.UpdateMarkRiskCAS(l.ctx, position, expectedVersion, utils.NowMillis())
			if err != nil {
				return err
			}
			if !updated {
				// A Fill, funding settlement or reservation changed the
				// position after this scan read it. The next scan recomputes
				// risk from the new position instead of overwriting it.
				continue
			}
		}
		if len(positions) < 100 {
			return nil
		}
	}
}

func markRiskProjectionEqual(before, after *models.TContractPosition) bool {
	return before != nil && after != nil &&
		before.MarkSnapshotId == after.MarkSnapshotId &&
		before.MarkPrice.Equal(after.MarkPrice) &&
		before.MaintenanceMargin.Equal(after.MaintenanceMargin) &&
		before.UnrealizedPnl.Equal(after.UnrealizedPnl) &&
		before.LiquidationPrice.Equal(after.LiquidationPrice) &&
		before.BankruptcyPrice.Equal(after.BankruptcyPrice) &&
		before.RiskRate.Equal(after.RiskRate) &&
		before.AdlRank == after.AdlRank
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
			if !position.Qty.IsPositive() || position.Status != int64(trade.PositionStatus_POSITION_STATUS_NORMAL) || !position.MarkPrice.IsPositive() || !position.LiquidationPrice.IsPositive() {
				continue
			}
			if !positionRequiresLiquidation(position) {
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
