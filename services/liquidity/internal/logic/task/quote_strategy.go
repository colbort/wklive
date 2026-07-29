package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"wklive/proto/common"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/shopspring/decimal"
)

type referenceQuote struct {
	price      decimal.Decimal
	source     string
	snapshotID string
	timestamp  int64
}

func prepareInternalQuoteCycles(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.LiquidityTaskReq) (int64, int64, error) {
	limit := int64(in.BatchSize)
	if limit <= 0 {
		limit = 100
	}
	configs, err := svcCtx.SymbolConfigModel.FindRunningInternal(ctx, in.ConfigId, limit)
	if err != nil {
		return 0, 0, err
	}
	var created, failed int64
	for _, config := range configs {
		ok, cycleErr := prepareInternalQuoteCycle(ctx, svcCtx, config)
		if cycleErr != nil {
			failed++
			continue
		}
		if ok {
			created++
		}
	}
	return created, failed, nil
}

func prepareInternalQuoteCycle(ctx context.Context, svcCtx *svc.ServiceContext, config *models.TLiquiditySymbolConfig) (bool, error) {
	now := time.Now().UnixMilli()
	latest, err := svcCtx.QuoteCycleModel.FindLatestByConfig(ctx, config.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return false, err
	}
	if latest != nil && config.RefreshIntervalMs > 0 && now-latest.StartedAt < config.RefreshIntervalMs {
		return false, nil
	}
	reference, err := loadReferenceQuote(ctx, svcCtx, config, now)
	if err != nil {
		return false, err
	}
	if latest != nil && latest.ReferencePrice.IsPositive() && config.RepriceThresholdBps.IsPositive() &&
		reference.price.Sub(latest.ReferencePrice).Abs().Div(latest.ReferencePrice).Mul(decimal.NewFromInt(10_000)).LessThan(config.RepriceThresholdBps) &&
		(config.QuoteTtlMs <= 0 || now-latest.StartedAt < config.QuoteTtlMs) {
		return false, nil
	}
	levels, err := svcCtx.StrategyLevelModel.FindList(ctx, config.Id, true)
	if err != nil {
		return false, err
	}
	if len(levels) == 0 {
		return false, fmt.Errorf("no enabled quote levels: config_id=%d", config.Id)
	}
	providerRow, err := svcCtx.ProviderModel.FindOne(ctx, config.InternalProviderId)
	if err != nil {
		return false, fmt.Errorf("internal provider is required: %w", err)
	}
	if err := svcCtx.InternalMarketMaker.Health(ctx, providerRow); err != nil {
		return false, err
	}
	orders := buildQuoteOrders(config, levels, reference.price, now)
	if len(orders) == 0 {
		return false, fmt.Errorf("no valid quote generated: config_id=%d", config.Id)
	}
	if err := cancelPreviousQuotes(ctx, svcCtx, config.Id); err != nil {
		return false, err
	}
	cycleNo := fmt.Sprintf("LQC%d-%d", config.Id, now)
	bidCount, askCount := countSides(orders)
	cycle := &models.TLiquidityQuoteCycle{
		CycleNo: cycleNo, ConfigId: config.Id, SymbolId: config.SymbolId,
		ReferencePrice: reference.price, ReferenceSource: reference.source,
		ReferenceSnapshotId: reference.snapshotID, ReferenceTime: reference.timestamp,
		TargetBidCount: bidCount, TargetAskCount: askCount,
		Status:    int64(liquidity.QuoteCycleStatus_QUOTE_CYCLE_STATUS_EXECUTING),
		StartedAt: now, CreateTimes: now, UpdateTimes: now,
	}
	result, err := svcCtx.QuoteCycleModel.Insert(ctx, cycle)
	if err != nil {
		return false, err
	}
	cycleID, err := result.LastInsertId()
	if err != nil {
		return false, err
	}
	for index, order := range orders {
		order.CycleId = cycleID
		order.QuoteNo = fmt.Sprintf("%s-Q%d", cycleNo, index+1)
		order.ClientOrderId = order.QuoteNo
		if _, err := svcCtx.QuoteOrderModel.Insert(ctx, order); err != nil {
			cycle.Status = int64(liquidity.QuoteCycleStatus_QUOTE_CYCLE_STATUS_FAILED)
			cycle.LastErrorMsg = err.Error()
			cycle.FinishedAt, cycle.UpdateTimes = time.Now().UnixMilli(), time.Now().UnixMilli()
			_ = svcCtx.QuoteCycleModel.Update(ctx, cycle)
			return false, err
		}
	}
	return true, nil
}

func loadReferenceQuote(ctx context.Context, svcCtx *svc.ServiceContext, config *models.TLiquiditySymbolConfig, targetTime int64) (*referenceQuote, error) {
	if svcCtx.MarketClient == nil {
		return nil, errors.New("market client is not configured")
	}
	kind := strings.ToUpper(strings.TrimSpace(config.ReferencePriceKind))
	if kind == "" || kind == "MARK_PRICE" {
		kind = "MARK"
	}
	authority := strings.TrimSpace(svcCtx.Config.PriceEngineAuthority)
	if kind == "FINAL_QUOTE" {
		authority = strings.TrimSpace(svcCtx.Config.MarketAuthority)
	}
	if authority == "" {
		return nil, errors.New("reference price authority is not configured")
	}
	validity := config.QuoteValidityMs
	if validity <= 0 {
		validity = 30_000
	}
	sources := strings.FieldsFunc(config.ReferencePriceSource, func(r rune) bool {
		return r == ',' || r == '|' || r == ';'
	})
	if len(sources) == 0 {
		sources = []string{config.ReferencePriceSource}
	}
	candidates := make([]*referenceQuote, 0, len(sources))
	for _, source := range sources {
		category, market, symbol := parseReferenceSource(source, config.Symbol)
		resp, err := svcCtx.MarketClient.GetAuthoritativeSnapshot(ctx, &market.GetAuthoritativeSnapshotReq{
			Authority: authority, CategoryCode: category, Market: market, Symbol: symbol,
			TargetTime: targetTime, MaxLookbackMs: validity, SnapshotKind: kind,
		})
		if err != nil || resp.GetData() == nil {
			continue
		}
		row := resp.GetData()
		price, parseErr := decimal.NewFromString(row.GetPrice())
		if parseErr != nil || !price.IsPositive() {
			continue
		}
		candidates = append(candidates, &referenceQuote{
			price: price, source: strings.TrimSpace(source), snapshotID: row.GetSnapshotId(),
			timestamp: row.GetSourceTimestamp(),
		})
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no valid reference price: source=%s", config.ReferencePriceSource)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].price.LessThan(candidates[j].price) })
	return candidates[len(candidates)/2], nil
}

func parseReferenceSource(source, fallbackSymbol string) (string, string, string) {
	parts := strings.Split(strings.TrimSpace(source), ":")
	category, market, symbol := "", "", fallbackSymbol
	if len(parts) > 0 {
		category = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 {
		market = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 && strings.TrimSpace(parts[2]) != "" {
		symbol = strings.TrimSpace(parts[2])
	}
	return category, market, symbol
}

func buildQuoteOrders(config *models.TLiquiditySymbolConfig, levels []*models.TLiquidityStrategyLevel, reference decimal.Decimal, now int64) []*models.TLiquidityQuoteOrder {
	orders := make([]*models.TLiquidityQuoteOrder, 0, len(levels)*2)
	expireAt := now + config.QuoteTtlMs
	if config.QuoteTtlMs <= 0 {
		expireAt = now + 30_000
	}
	for _, level := range levels {
		bidSpread := clampSpread(config.BaseSpreadBps.Add(level.BidSpreadBps), config)
		askSpread := clampSpread(config.BaseSpreadBps.Add(level.AskSpreadBps), config)
		bidPrice := roundDown(reference.Mul(decimal.NewFromInt(1).Sub(bidSpread.Div(decimal.NewFromInt(10_000)))), config.PriceTick)
		askPrice := roundUp(reference.Mul(decimal.NewFromInt(1).Add(askSpread.Div(decimal.NewFromInt(10_000)))), config.PriceTick)
		if bidPrice.IsPositive() && bidPrice.LessThan(reference) {
			if qty := normalizeQty(level.BidQty, bidPrice, config); qty.IsPositive() {
				orders = append(orders, newQuoteOrder(config, level.LevelNo, int64(common.Side_SIDE_BUY), bidPrice, qty, expireAt, now))
			}
		}
		if askPrice.GreaterThan(reference) {
			if qty := normalizeQty(level.AskQty, askPrice, config); qty.IsPositive() {
				orders = append(orders, newQuoteOrder(config, level.LevelNo, int64(common.Side_SIDE_SELL), askPrice, qty, expireAt, now))
			}
		}
	}
	return orders
}

func newQuoteOrder(config *models.TLiquiditySymbolConfig, level, side int64, price, qty decimal.Decimal, expireAt, now int64) *models.TLiquidityQuoteOrder {
	return &models.TLiquidityQuoteOrder{
		ConfigId: config.Id, ProviderId: config.InternalProviderId, SymbolId: config.SymbolId,
		Side: side, LevelNo: level, Price: price, Qty: qty,
		Status:   int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT),
		ExpireAt: expireAt, Version: 1, CreateTimes: now, UpdateTimes: now,
	}
}

func clampSpread(spread decimal.Decimal, config *models.TLiquiditySymbolConfig) decimal.Decimal {
	if spread.IsNegative() {
		spread = decimal.Zero
	}
	limit := config.MaxSpreadBps
	if config.MaxPriceDeviationBps.IsPositive() && (!limit.IsPositive() || config.MaxPriceDeviationBps.LessThan(limit)) {
		limit = config.MaxPriceDeviationBps
	}
	if limit.IsPositive() && spread.GreaterThan(limit) {
		return limit
	}
	return spread
}

func normalizeQty(qty, price decimal.Decimal, config *models.TLiquiditySymbolConfig) decimal.Decimal {
	qty = roundDown(qty, config.QtyStep)
	if config.MaxQuoteQty.IsPositive() && qty.GreaterThan(config.MaxQuoteQty) {
		qty = roundDown(config.MaxQuoteQty, config.QtyStep)
	}
	if config.MaxQuoteNotional.IsPositive() && qty.Mul(price).GreaterThan(config.MaxQuoteNotional) {
		qty = roundDown(config.MaxQuoteNotional.Div(price), config.QtyStep)
	}
	if !qty.IsPositive() || (config.MinQuoteQty.IsPositive() && qty.LessThan(config.MinQuoteQty)) {
		return decimal.Zero
	}
	return qty
}

func roundDown(value, step decimal.Decimal) decimal.Decimal {
	if !step.IsPositive() {
		return value
	}
	return value.Div(step).Floor().Mul(step)
}

func roundUp(value, step decimal.Decimal) decimal.Decimal {
	if !step.IsPositive() {
		return value
	}
	return value.Div(step).Ceil().Mul(step)
}

func countSides(orders []*models.TLiquidityQuoteOrder) (int64, int64) {
	var bids, asks int64
	for _, order := range orders {
		if order.Side == int64(common.Side_SIDE_BUY) {
			bids++
		} else if order.Side == int64(common.Side_SIDE_SELL) {
			asks++
		}
	}
	return bids, asks
}

func cancelPreviousQuotes(ctx context.Context, svcCtx *svc.ServiceContext, configID int64) error {
	rows, err := svcCtx.QuoteOrderModel.FindActiveByConfig(ctx, configID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Status == int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT) {
			row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED)
			row.CancelReason = "replaced by next quote cycle"
			row.Version++
			row.UpdateTimes = time.Now().UnixMilli()
			if err := svcCtx.QuoteOrderModel.Update(ctx, row); err != nil {
				return err
			}
			continue
		}
		providerRow, err := svcCtx.ProviderModel.FindOne(ctx, row.ProviderId)
		if err != nil {
			return err
		}
		result, err := svcCtx.InternalMarketMaker.CancelQuote(ctx, providerRow, row)
		if err != nil {
			return err
		}
		applyQuoteResult(row, result)
		if err := svcCtx.QuoteOrderModel.Update(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
