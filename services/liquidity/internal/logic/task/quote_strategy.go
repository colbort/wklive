package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"wklive/proto/common"
	"wklive/proto/liquidity"
	pb "wklive/proto/market"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

type referenceQuote struct {
	price      decimal.Decimal
	source     string
	snapshotID string
	timestamp  int64
}

const (
	tradingStatusLogInterval    = 5 * time.Minute
	referenceRetryInitial       = 5 * time.Second
	referenceRetryMaximum       = 30 * time.Second
	referenceFailureLogInterval = 5 * time.Minute
)

type tradingStatusLogEntry struct {
	state    string
	loggedAt time.Time
}

var tradingStatusLogs = struct {
	sync.Mutex
	entries map[int64]tradingStatusLogEntry
}{entries: make(map[int64]tradingStatusLogEntry)}

var uncertainQuoteLogs = struct {
	sync.Mutex
	entries map[int64]time.Time
}{entries: make(map[int64]time.Time)}

type referenceRetryEntry struct {
	failures int
	retryAt  time.Time
	loggedAt time.Time
}

var referenceRetries = struct {
	sync.Mutex
	entries map[int64]referenceRetryEntry
}{entries: make(map[int64]referenceRetryEntry)}

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
	open, err := ensureMarketOpen(ctx, svcCtx, config, now)
	if err != nil {
		return false, err
	}
	if !open {
		clearReferenceRetry(config.Id)
		return false, nil
	}
	hasUncertain, err := svcCtx.QuoteOrderModel.HasUncertainByConfig(ctx, config.Id)
	if err != nil {
		return false, err
	}
	if hasUncertain {
		logUncertainQuoteBackpressure(ctx, config)
		return false, nil
	}
	latest, err := svcCtx.QuoteCycleModel.FindLatestByConfig(ctx, config.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return false, err
	}
	if latest != nil && config.RefreshIntervalMs > 0 && now-latest.StartedAt < config.RefreshIntervalMs {
		return false, nil
	}
	if !referenceRetryAllowed(config.Id, time.UnixMilli(now)) {
		return false, nil
	}
	reference, err := loadReferenceQuoteOrCancel(ctx, svcCtx, config, now)
	if err != nil {
		recordReferenceFailure(ctx, config, err)
		return false, err
	}
	clearReferenceRetry(config.Id)
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
	if err := cancelActiveQuotes(ctx, svcCtx, config.Id, "replaced by next quote cycle"); err != nil {
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

func logUncertainQuoteBackpressure(ctx context.Context, config *models.TLiquiditySymbolConfig) {
	now := time.Now()
	uncertainQuoteLogs.Lock()
	last := uncertainQuoteLogs.entries[config.Id]
	if now.Sub(last) < tradingStatusLogInterval {
		uncertainQuoteLogs.Unlock()
		return
	}
	uncertainQuoteLogs.entries[config.Id] = now
	uncertainQuoteLogs.Unlock()
	logx.WithContext(ctx).Errorf(
		"skip liquidity quote cycle while prior quote outcome is uncertain: config_id=%d symbol_id=%d symbol=%s",
		config.Id, config.SymbolId, config.Symbol,
	)
}

func loadReferenceQuoteOrCancel(ctx context.Context, svcCtx *svc.ServiceContext, config *models.TLiquiditySymbolConfig, targetTime int64) (*referenceQuote, error) {
	reference, err := loadReferenceQuote(ctx, svcCtx, config, targetTime)
	if err == nil {
		return reference, nil
	}
	cancelErr := cancelActiveQuotes(ctx, svcCtx, config.Id, "reference price unavailable")
	if cancelErr != nil {
		return nil, errors.Join(err, cancelErr)
	}
	return nil, err
}

func referenceRetryAllowed(configID int64, now time.Time) bool {
	referenceRetries.Lock()
	entry, found := referenceRetries.entries[configID]
	referenceRetries.Unlock()
	return !found || !now.Before(entry.retryAt)
}

func recordReferenceFailure(ctx context.Context, config *models.TLiquiditySymbolConfig, referenceErr error) {
	if config == nil {
		return
	}
	now := time.Now()
	referenceRetries.Lock()
	entry := referenceRetries.entries[config.Id]
	entry.failures++
	delay := referenceRetryInitial
	for i := 1; i < entry.failures && delay < referenceRetryMaximum; i++ {
		delay *= 2
		if delay > referenceRetryMaximum {
			delay = referenceRetryMaximum
		}
	}
	entry.retryAt = now.Add(delay)
	shouldLog := entry.loggedAt.IsZero() || now.Sub(entry.loggedAt) >= referenceFailureLogInterval
	if shouldLog {
		entry.loggedAt = now
	}
	referenceRetries.entries[config.Id] = entry
	referenceRetries.Unlock()
	if shouldLog {
		logx.WithContext(ctx).Errorf(
			"[LIQUIDITY_REFERENCE_PRICE] config_id=%d symbol=%s source=%s unavailable=true failures=%d retry_in_ms=%d err=%v",
			config.Id, config.Symbol, config.ReferencePriceSource, entry.failures, delay.Milliseconds(), referenceErr,
		)
	}
}

func clearReferenceRetry(configID int64) {
	referenceRetries.Lock()
	delete(referenceRetries.entries, configID)
	referenceRetries.Unlock()
}

func loadReferenceQuote(ctx context.Context, svcCtx *svc.ServiceContext, config *models.TLiquiditySymbolConfig, targetTime int64) (*referenceQuote, error) {
	if svcCtx.MarketClient == nil {
		return nil, errors.New("market client is not configured")
	}
	kind := strings.ToUpper(strings.TrimSpace(config.ReferencePriceKind))
	if kind == "" || kind == "MARK_PRICE" {
		kind = "MARK"
	}
	var authorities []string
	if kind == "FINAL_QUOTE" {
		authorities = normalizeAuthorities(svcCtx.Config.MarketAuthorities)
	} else {
		authority := strings.ToLower(strings.TrimSpace(svcCtx.Config.PriceEngineAuthority))
		if authority != "" {
			authorities = []string{authority}
		}
	}
	if len(authorities) == 0 {
		return nil, errors.New("reference price authority is not configured")
	}
	validity := config.QuoteValidityMs
	if validity <= 0 {
		validity = 30_000
	}
	sources := referencePriceSources(config.ReferencePriceSource)
	if len(sources) == 0 {
		sources = []string{config.ReferencePriceSource}
	}
	candidates := make([]*referenceQuote, 0, len(sources))
	for _, source := range sources {
		category, market, symbol := parseReferenceSource(source, config.Symbol)
		for _, candidateAuthority := range authorities {
			resp, err := svcCtx.MarketClient.GetAuthoritativeSnapshot(ctx, &pb.GetAuthoritativeSnapshotReq{
				Authority: candidateAuthority, CategoryCode: category, Market: market, Symbol: symbol,
				TargetTime: targetTime, MaxLookbackMs: validity, SnapshotKind: kind,
			})
			if err != nil || resp.GetData() == nil {
				continue
			}
			row := resp.GetData()
			if !referenceSnapshotFresh(targetTime, row.GetSourceTimestamp(), validity) {
				continue
			}
			price, parseErr := decimal.NewFromString(row.GetPrice())
			if parseErr != nil || !price.IsPositive() {
				continue
			}
			candidates = append(candidates, &referenceQuote{
				price: price, source: strings.TrimSpace(source), snapshotID: row.GetSnapshotId(),
				timestamp: row.GetSourceTimestamp(),
			})
			break
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no valid reference price: source=%s", config.ReferencePriceSource)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].price.LessThan(candidates[j].price) })
	return candidates[len(candidates)/2], nil
}

func normalizeAuthorities(values []string) []string {
	authorities := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		for _, candidate := range strings.FieldsFunc(value, func(r rune) bool {
			return r == ',' || r == '|' || r == ';'
		}) {
			candidate = strings.ToLower(strings.TrimSpace(candidate))
			if candidate == "" {
				continue
			}
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}
			authorities = append(authorities, candidate)
		}
	}
	return authorities
}

func referenceSnapshotFresh(targetTime, sourceTimestamp, validity int64) bool {
	if targetTime <= 0 || sourceTimestamp <= 0 || validity <= 0 {
		return false
	}
	age := targetTime - sourceTimestamp
	return age >= 0 && age <= validity
}

func referencePriceSources(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '|' || r == ';'
	})
}

func ensureMarketOpen(ctx context.Context, svcCtx *svc.ServiceContext, config *models.TLiquiditySymbolConfig, timestamp int64) (bool, error) {
	open, reason, sessionType, err := loadTradingStatus(ctx, svcCtx, config, timestamp)
	if err != nil {
		logTradingStatus(ctx, config, false, "market_status_unavailable", sessionType, err)
		cancelErr := cancelActiveQuotes(ctx, svcCtx, config.Id, "market status unavailable")
		if cancelErr != nil {
			return false, errors.Join(err, cancelErr)
		}
		return false, err
	}
	category, _, _ := parseReferenceSource(config.ReferencePriceSource, config.Symbol)
	if open && strings.EqualFold(category, "stock") && !strings.EqualFold(sessionType, "regular") {
		open = false
		if strings.TrimSpace(sessionType) == "" {
			reason = "stock_session_type_unavailable"
		} else {
			reason = "stock_extended_session_disabled:" + strings.ToLower(strings.TrimSpace(sessionType))
		}
	}
	if open {
		logTradingStatus(ctx, config, true, reason, sessionType, nil)
		return true, nil
	}
	if reason == "" {
		reason = "market_closed"
	}
	logTradingStatus(ctx, config, false, reason, sessionType, nil)
	if err := cancelActiveQuotes(ctx, svcCtx, config.Id, reason); err != nil {
		return false, err
	}
	return false, nil
}

func logTradingStatus(ctx context.Context, config *models.TLiquiditySymbolConfig, open bool, reason, sessionType string, statusErr error) {
	if config == nil {
		return
	}
	state := fmt.Sprintf("%t:%s:%s", open, reason, sessionType)
	now := time.Now()
	tradingStatusLogs.Lock()
	previous, found := tradingStatusLogs.entries[config.Id]
	if found && previous.state == state && now.Sub(previous.loggedAt) < tradingStatusLogInterval {
		tradingStatusLogs.Unlock()
		return
	}
	tradingStatusLogs.entries[config.Id] = tradingStatusLogEntry{state: state, loggedAt: now}
	tradingStatusLogs.Unlock()

	source := strings.TrimSpace(config.ReferencePriceSource)
	if sources := referencePriceSources(source); len(sources) > 0 {
		source = sources[0]
	}
	logger := logx.WithContext(ctx)
	if statusErr != nil {
		logger.Errorf("[LIQUIDITY_MARKET_STATUS] config_id=%d symbol=%s source=%s open=false reason=%s session_type=%s err=%v", config.Id, config.Symbol, source, reason, sessionType, statusErr)
		return
	}
	if !open {
		logger.Errorf("[LIQUIDITY_MARKET_STATUS] config_id=%d symbol=%s source=%s open=false reason=%s session_type=%s", config.Id, config.Symbol, source, reason, sessionType)
		return
	}
	logger.Infof("[LIQUIDITY_MARKET_STATUS] config_id=%d symbol=%s source=%s open=%t reason=%s session_type=%s", config.Id, config.Symbol, source, open, reason, sessionType)
}

func loadTradingStatus(ctx context.Context, svcCtx *svc.ServiceContext, config *models.TLiquiditySymbolConfig, timestamp int64) (bool, string, string, error) {
	if svcCtx.MarketClient == nil {
		return false, "", "", errors.New("market client is not configured")
	}
	sources := referencePriceSources(config.ReferencePriceSource)
	if len(sources) == 0 {
		return false, "", "", errors.New("reference price source is required for trading calendar")
	}
	category, marketCode, symbol := parseReferenceSource(sources[0], config.Symbol)
	if category == "" || marketCode == "" || symbol == "" {
		return false, "", "", fmt.Errorf("invalid trading calendar source: %s", sources[0])
	}
	resp, err := svcCtx.MarketClient.GetTradingStatus(ctx, &pb.GetTradingStatusReq{
		CategoryCode: category,
		Market:       marketCode,
		Symbol:       symbol,
		Timestamp:    timestamp,
	})
	if err != nil {
		return false, "", "", fmt.Errorf("get market trading status: %w", err)
	}
	if resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		return false, "", "", fmt.Errorf("market trading status unavailable: %s", resp.GetBase().GetMsg())
	}
	return resp.GetData().GetIsOpen(), resp.GetData().GetReason(), resp.GetData().GetSessionType(), nil
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

func cancelActiveQuotes(ctx context.Context, svcCtx *svc.ServiceContext, configID int64, reason string) error {
	rows, err := svcCtx.QuoteOrderModel.FindActiveByConfig(ctx, configID)
	if err != nil {
		return err
	}
	if len(rows) > 0 {
		logx.WithContext(ctx).Infof("cancel active liquidity quotes: config_id=%d count=%d reason=%s", configID, len(rows), reason)
	}
	for _, row := range rows {
		if row.Status == int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT) {
			row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED)
			row.CancelReason = reason
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
		if row.Status == int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED) ||
			row.Status == int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELING) {
			row.CancelReason = reason
		}
		if err := svcCtx.QuoteOrderModel.Update(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
