package adminlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"wklive/services/trade/internal/logic/helpers"

	cache "wklive/common/market"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	pb "wklive/proto/market"
	"wklive/proto/trade"
	"wklive/services/trade/internal/domain/contractmath"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

const secondsWorkBatchSize = int64(100)
const secondsWorkLeaseMillis = int64(60_000)

type ProcessSecondsSettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type marketQuoteSnapshot struct {
	Category       string `json:"category"`
	Market         string `json:"market"`
	Symbol         string `json:"symbol"`
	LastPrice      string `json:"last_price"`
	OpenPrice      string `json:"open_price"`
	HighPrice      string `json:"high_price"`
	LowPrice       string `json:"low_price"`
	Volume         string `json:"volume"`
	Turnover       string `json:"turnover"`
	QuoteTs        int64  `json:"quote_ts"`
	ReceivedAt     int64  `json:"received_at"`
	SnapshotID     string `json:"snapshot_id"`
	Revision       int64  `json:"revision"`
	FormulaVersion string `json:"formula_version"`
	Confirmed      bool   `json:"confirmed"`
}

func secondsWorkLeaseOwned(current *models.TTradeOrderSeconds, status trade.SecondsSettlementStatus, lease int64) bool {
	return current != nil && current.SettlementStatus == int64(status) && current.UpdateTimes == lease
}

func NewProcessSecondsSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessSecondsSettlementsLogic {
	return &ProcessSecondsSettlementsLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ProcessSecondsSettlementsLogic) Process(tenantID int64) error {
	return runSecondsPhases(
		func() error { return l.processActivations(tenantID) },
		func() error { return l.processSettlements(tenantID) },
		func() error { return l.processRefunds(tenantID) },
	)
}

func runSecondsPhases(phases ...func() error) error {
	var result error
	for _, phase := range phases {
		if phase != nil {
			result = errors.Join(result, phase())
		}
	}
	return result
}

func (l *ProcessSecondsSettlementsLogic) processActivations(tenantID int64) error {
	return l.scan(tenantID, int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVATING), 0, func(item *models.SecondsOrderWorkItem) error {
		now := utils.NowMillis()
		claimed, lease, err := l.svcCtx.TradeOrderSecondsModel.ClaimActivation(l.ctx, item.Id, now, now-secondsWorkLeaseMillis)
		if err != nil || !claimed {
			return err
		}
		item.UpdateTimes = lease
		cfg, err := l.svcCtx.TradeSymbolSecondsModel.FindOneByTenantIdSymbolIdDurationSeconds(l.ctx, item.TenantId, item.SymbolId, item.DurationSeconds)
		if err != nil {
			return err
		}
		quote, candidates, err := l.getValidQuotesKind("SECONDS_START", cfg.StartPriceSource, item.SymbolId, cfg.QuoteValidityMs)
		if err != nil {
			return l.moveSecondsToRefund(item, "invalid start quote: "+err.Error())
		}
		now = utils.NowMillis()
		return l.svcCtx.TransactionModel.TransactOnce(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
			model := tx.TradeOrderSeconds
			current, err := model.FindOneForUpdate(ctx, item.Id)
			if err != nil {
				return err
			}
			if !secondsWorkLeaseOwned(current, trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVATING, item.UpdateTimes) {
				return errors.New("seconds activation lease lost")
			}
			current.ActivatedAt, current.StartPriceTime, current.ExpireTime = now, quote.QuoteTs, now+current.DurationSeconds*1000
			current.StartPrice = helpers.MustParseFloat(quote.LastPrice)
			current.StartPriceSource = quoteSource(quote)
			current.PriceAlgorithm = nonEmpty(cfg.SettlementPriceAlgorithm, "last-v1")
			current.SettlementStatus = int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVE)
			current.RetryCount, current.NextRetryAt, current.LastErrorMsg = 0, 0, ""
			current.Version++
			current.UpdateTimes = now
			if err := model.Update(ctx, current); err != nil {
				return err
			}
			for _, candidate := range candidates {
				if err := insertSecondsPriceSnapshot(ctx, tx, current, candidate, trade.SecondsPriceSnapshotType_SECONDS_PRICE_SNAPSHOT_TYPE_START, candidate == quote); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

func (l *ProcessSecondsSettlementsLogic) processSettlements(tenantID int64) error {
	for _, status := range []trade.SecondsSettlementStatus{trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVE, trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING} {
		if err := l.scan(tenantID, int64(status), utils.NowMillis(), func(item *models.SecondsOrderWorkItem) error {
			now := utils.NowMillis()
			claimed, lease, err := l.svcCtx.TradeOrderSecondsModel.ClaimSettlement(l.ctx, item.Id, now, now-secondsWorkLeaseMillis)
			if err != nil || !claimed {
				return err
			}
			item.SettlementStatus = int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING)
			item.UpdateTimes = lease
			cfg, err := l.svcCtx.TradeSymbolSecondsModel.FindOneByTenantIdSymbolIdDurationSeconds(l.ctx, item.TenantId, item.SymbolId, item.DurationSeconds)
			if err != nil {
				return err
			}
			quote, candidates, err := l.getValidQuotesAtKind("SECONDS_SETTLEMENT", cfg.SettlementPriceSource, item.SymbolId, item.ExpireTime, cfg.QuoteValidityMs)
			if err != nil {
				return l.moveSecondsToRefund(item, "invalid settlement quote: "+err.Error())
			}
			if window := cfg.SettlementWindowMs; window > 0 && (quote.QuoteTs < item.ExpireTime-window || quote.QuoteTs > item.ExpireTime+window) {
				return l.moveSecondsToRefund(item, "settlement quote outside configured window")
			}
			price := helpers.MustParseFloat(quote.LastPrice)
			result := secondsResult(item.Direction, item.StartPrice, price, cfg.DrawTolerance)
			if result == trade.SecondsResult_SECONDS_RESULT_DRAW && cfg.DrawRule == int64(trade.SecondsDrawRule_SECONDS_DRAW_RULE_LOSE) {
				result = trade.SecondsResult_SECONDS_RESULT_LOSE
			}
			if result == trade.SecondsResult_SECONDS_RESULT_DRAW {
				return l.moveSecondsToRefund(item, "draw refund")
			}
			profit, fee, returned := secondsPayout(item.StakeAmount, item.PayoutRate, item.FeeRate, result)
			if err := l.consumeSecondsStake(item); err != nil {
				return err
			}
			if returned.IsPositive() {
				if err := l.creditSeconds(item, returned); err != nil {
					return err
				}
			}
			now = utils.NowMillis()
			return l.svcCtx.TransactionModel.TransactOnce(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
				secondsModel := tx.TradeOrderSeconds
				orderModel := tx.TradeOrder
				eventModel := tx.TradeEventOutbox
				current, err := secondsModel.FindOneForUpdate(ctx, item.Id)
				if err != nil {
					return err
				}
				if current.SettlementStatus == int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLED) {
					return nil
				}
				if !secondsWorkLeaseOwned(current, trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING, item.UpdateTimes) {
					return errors.New("seconds settlement lease lost")
				}
				current.SettlementStatus = int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLED)
				current.Result = int64(result)
				current.SettlementPrice = price
				current.SettlementPriceTime = quote.QuoteTs
				current.SettlementPriceSource = quoteSource(quote)
				current.ProfitAmount = profit
				current.FeeAmount = fee
				current.ReturnAmount = returned
				current.SettlementReason = "settled"
				current.SettledAt = now
				current.RetryCount, current.NextRetryAt, current.LastErrorMsg = 0, 0, ""
				current.Version++
				current.UpdateTimes = now
				if err := secondsModel.Update(ctx, current); err != nil {
					return err
				}
				order, err := orderModel.FindOneForUpdate(ctx, current.OrderId)
				if err != nil {
					return err
				}
				order.Status = int64(trade.OrderStatus_ORDER_STATUS_FILLED)
				order.Version++
				order.UpdateTimes = now
				if err := orderModel.Update(ctx, order); err != nil {
					return err
				}
				if err := helpers.InsertOrderChangedOutbox(ctx, eventModel, order, order.OrderNo+"-SECONDS-SETTLED", "ORDER_SETTLED", now); err != nil {
					return err
				}
				for _, candidate := range candidates {
					if err := insertSecondsPriceSnapshot(ctx, tx, current, candidate, trade.SecondsPriceSnapshotType_SECONDS_PRICE_SNAPSHOT_TYPE_SETTLEMENT_CANDIDATE, false); err != nil {
						return err
					}
				}
				return insertSecondsPriceSnapshot(ctx, tx, current, quote, trade.SecondsPriceSnapshotType_SECONDS_PRICE_SNAPSHOT_TYPE_FINAL_SETTLEMENT, true)
			})
		}); err != nil {
			return err
		}
	}
	return nil
}

func (l *ProcessSecondsSettlementsLogic) processRefunds(tenantID int64) error {
	return l.scan(tenantID, int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_REFUNDING), 0, func(item *models.SecondsOrderWorkItem) error {
		now := utils.NowMillis()
		claimed, lease, err := l.svcCtx.TradeOrderSecondsModel.ClaimRefund(l.ctx, item.Id, now, now-secondsWorkLeaseMillis)
		if err != nil || !claimed {
			return err
		}
		item.UpdateTimes = lease
		resp, err := l.svcCtx.AssetClient.UnfreezeAssetByBizNo(l.ctx, &asset.UnfreezeAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.OrderNo, Amount: item.StakeAmount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: item.Id, BizNo: item.OrderNo + "-SECONDS-REFUND", Remark: "seconds contract refund"})
		if err != nil {
			return err
		}
		if err = validateSecondsAssetResponse("refund", resp); err != nil {
			return err
		}
		now = utils.NowMillis()
		return l.svcCtx.TransactionModel.TransactOnce(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
			secondsModel := tx.TradeOrderSeconds
			orderModel := tx.TradeOrder
			eventModel := tx.TradeEventOutbox
			current, err := secondsModel.FindOneForUpdate(ctx, item.Id)
			if err != nil {
				return err
			}
			if current.SettlementStatus == int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_REFUNDED) {
				return nil
			}
			if !secondsWorkLeaseOwned(current, trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_REFUNDING, item.UpdateTimes) {
				return errors.New("seconds refund lease lost")
			}
			current.SettlementStatus = int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_REFUNDED)
			current.Result = int64(trade.SecondsResult_SECONDS_RESULT_VOID)
			current.ReturnAmount = current.StakeAmount
			current.SettledAt, current.Version, current.UpdateTimes = now, current.Version+1, now
			current.RetryCount, current.NextRetryAt, current.LastErrorMsg = 0, 0, ""
			if err = secondsModel.Update(ctx, current); err != nil {
				return err
			}
			order, err := orderModel.FindOneForUpdate(ctx, current.OrderId)
			if err != nil {
				return err
			}
			order.Status, order.CancelReason, order.Version, order.UpdateTimes = int64(trade.OrderStatus_ORDER_STATUS_CANCELED), current.SettlementReason, order.Version+1, now
			if err := orderModel.Update(ctx, order); err != nil {
				return err
			}
			return helpers.InsertOrderChangedOutbox(ctx, eventModel, order, order.OrderNo+"-SECONDS-REFUNDED", "ORDER_CANCELED", now)
		})
	})
}

func (l *ProcessSecondsSettlementsLogic) scan(tenantID, status, due int64, fn func(*models.SecondsOrderWorkItem) error) error {
	return scanSecondsWork(func(cursor int64) ([]*models.SecondsOrderWorkItem, error) {
		return l.svcCtx.TradeOrderSecondsModel.FindWork(l.ctx, tenantID, status, due, cursor, secondsWorkBatchSize)
	}, func(item *models.SecondsOrderWorkItem) error {
		itemErr := fn(item)
		if itemErr == nil {
			return nil
		}
		_, markErr := l.svcCtx.TradeOrderSecondsModel.MarkWorkFailure(l.ctx, item.Id, item.SettlementStatus, item.UpdateTimes, itemErr.Error(), utils.NowMillis())
		return errors.Join(itemErr, markErr)
	})
}

func scanSecondsWork(fetch func(int64) ([]*models.SecondsOrderWorkItem, error), fn func(*models.SecondsOrderWorkItem) error) error {
	cursor := int64(0)
	var firstErr error
	for {
		items, err := fetch(cursor)
		if err != nil {
			if firstErr != nil {
				return fmt.Errorf("seconds work item failed: %v; scan failed: %w", firstErr, err)
			}
			return err
		}
		if len(items) == 0 {
			return firstErr
		}
		for _, item := range items {
			cursor = item.Id
			if itemErr := fn(item); itemErr != nil && firstErr == nil {
				firstErr = fmt.Errorf("seconds order id=%d: %w", item.Id, itemErr)
			}
		}
		if int64(len(items)) < secondsWorkBatchSize {
			return firstErr
		}
	}
}

func (l *ProcessSecondsSettlementsLogic) moveSecondsToRefund(item *models.SecondsOrderWorkItem, reason string) error {
	return l.svcCtx.TransactionModel.TransactOnce(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		model := tx.TradeOrderSeconds
		current, err := model.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if current.SettlementStatus == int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_REFUNDING) || current.SettlementStatus == int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_REFUNDED) {
			return nil
		}
		if item.SettlementStatus == int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING) {
			if !secondsWorkLeaseOwned(current, trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING, item.UpdateTimes) {
				return errors.New("seconds settlement lease lost before refund")
			}
		} else if !secondsWorkLeaseOwned(current, trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVATING, item.UpdateTimes) {
			return errors.New("seconds activation lease lost before refund")
		}
		current.SettlementStatus = int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_REFUNDING)
		current.Result = int64(trade.SecondsResult_SECONDS_RESULT_VOID)
		current.SettlementReason = reason
		current.NextRetryAt, current.LastErrorMsg = 0, ""
		current.Version++
		current.UpdateTimes = 0
		return model.Update(ctx, current)
	})
}

func (l *ProcessSecondsSettlementsLogic) getValidQuote(source string, symbolID, validity int64) (*marketQuoteSnapshot, error) {
	return l.getValidQuoteKind("SETTLEMENT_PRICE", source, symbolID, validity)
}
func (l *ProcessSecondsSettlementsLogic) getValidQuoteKind(kind, source string, symbolID, validity int64) (*marketQuoteSnapshot, error) {
	selected, _, err := l.getValidQuotesKind(kind, source, symbolID, validity)
	return selected, err
}

func (l *ProcessSecondsSettlementsLogic) getValidQuotes(source string, symbolID, validity int64) (*marketQuoteSnapshot, []*marketQuoteSnapshot, error) {
	return l.getValidQuotesKind("SETTLEMENT_PRICE", source, symbolID, validity)
}
func (l *ProcessSecondsSettlementsLogic) getValidQuotesKind(kind, source string, symbolID, validity int64) (*marketQuoteSnapshot, []*marketQuoteSnapshot, error) {
	return l.getValidQuotesAtKind(kind, source, symbolID, utils.NowMillis(), validity)
}

func (l *ProcessSecondsSettlementsLogic) getValidQuotesAtKind(kind, source string, symbolID, targetTime, validity int64) (*marketQuoteSnapshot, []*marketQuoteSnapshot, error) {
	sources := strings.FieldsFunc(source, func(r rune) bool { return r == ',' || r == '|' || r == ';' })
	if len(sources) == 0 {
		sources = []string{source}
	}
	candidates := make([]*marketQuoteSnapshot, 0, len(sources))
	for _, currentSource := range sources {
		q, err := l.getOneValidQuote(kind, strings.TrimSpace(currentSource), symbolID, targetTime, validity)
		if err == nil {
			candidates = append(candidates, q)
		}
	}
	if len(candidates) == 0 {
		return nil, nil, fmt.Errorf("no valid market quote: source=%s", source)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return helpers.MustParseFloat(candidates[i].LastPrice).LessThan(helpers.MustParseFloat(candidates[j].LastPrice))
	})
	return candidates[len(candidates)/2], candidates, nil
}

func (l *ProcessSecondsSettlementsLogic) getOneValidQuote(kind, source string, symbolID, targetTime, validity int64) (*marketQuoteSnapshot, error) {
	category, market, symbol := parseQuoteSource(source)
	tradeSymbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, symbolID)
	if err != nil {
		return nil, err
	}
	if symbol == "" {
		symbol = tradeSymbol.Symbol
	}
	if validity <= 0 {
		validity = 30_000
	}
	snapshotKind := archiveSnapshotKind(kind)
	authority := strings.TrimSpace(l.svcCtx.Config.MarketAuthority)
	if snapshotKind != "FINAL_QUOTE" {
		authority = strings.TrimSpace(l.svcCtx.Config.PriceEngineAuthority)
	}
	if authority == "" {
		return nil, errors.New("market authority is not configured")
	}
	s, err := l.svcCtx.MarketDataCache.FindAuthoritativeSnapshotAt(l.ctx, cache.ClientMessage{Topic: cache.TopicQuote, CategoryCode: category, Market: market, Symbol: symbol}, authority, snapshotKind, targetTime, time.Duration(validity)*time.Millisecond)
	if err != nil {
		s, err = l.findAuthoritativeSnapshotInArchive(authority, snapshotKind, category, market, symbol, targetTime, validity)
		if err != nil {
			return nil, err
		}
	}
	if err = persistMarketSnapshot(l.ctx, l.svcCtx.TradeMarketSnapshotModel, tradeSymbol.TenantId, symbolID, s); err != nil {
		return nil, err
	}
	q := &marketQuoteSnapshot{Category: category, Market: market, Symbol: symbol, LastPrice: s.Price, QuoteTs: s.SourceTimestamp, ReceivedAt: s.SnapshotTimestamp, SnapshotID: s.SnapshotID, Revision: s.Revision, FormulaVersion: s.FormulaVersion, Confirmed: s.Confirmed}
	if quoteIsValidAtKind(q, targetTime, validity, snapshotKind) {
		return q, nil
	}
	return nil, fmt.Errorf("market quote cache miss: source=%s", source)
}

func (l *ProcessSecondsSettlementsLogic) findAuthoritativeSnapshotInArchive(authority, kind, category, market, symbol string, targetTime, validity int64) (*cache.SettlementSnapshot, error) {
	if l.svcCtx.MarketClient == nil {
		return nil, errors.New("market archive client is not configured")
	}
	resp, err := l.svcCtx.MarketClient.GetAuthoritativeSnapshot(l.ctx, &pb.GetAuthoritativeSnapshotReq{
		Authority:     authority,
		SnapshotKind:  kind,
		CategoryCode:  category,
		Market:        market,
		Symbol:        symbol,
		TargetTime:    targetTime,
		MaxLookbackMs: validity,
	})
	if err != nil || resp.GetData() == nil {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("authoritative snapshot unavailable in archive")
	}
	row := resp.GetData()
	return &cache.SettlementSnapshot{
		SnapshotID:        row.GetSnapshotId(),
		Authority:         row.GetAuthority(),
		Kind:              row.GetSnapshotKind(),
		CategoryCode:      row.GetCategoryCode(),
		Market:            row.GetMarket(),
		Symbol:            row.GetSymbol(),
		Price:             row.GetPrice(),
		Source:            row.GetAuthority(),
		SourceTimestamp:   row.GetSourceTimestamp(),
		SnapshotTimestamp: row.GetSnapshotTimestamp(),
		Revision:          row.GetRevision(),
		FormulaVersion:    row.GetFormulaVersion(),
		Confirmed:         true,
	}, nil
}

func archiveSnapshotKind(kind string) string {
	switch strings.ToUpper(strings.TrimSpace(kind)) {
	case "MARK_PRICE", "MARK":
		return "MARK"
	case "INDEX_PRICE", "INDEX":
		return "INDEX"
	case "FUNDING_RATE", "FUNDING":
		return "FUNDING"
	case "DELIVERY_PRICE", "DELIVERY":
		return "DELIVERY"
	default:
		return "FINAL_QUOTE"
	}
}

func quoteIsValidAtKind(q *marketQuoteSnapshot, targetTime, validity int64, kind string) bool {
	if q == nil || !q.Confirmed || q.SnapshotID == "" || q.QuoteTs <= 0 || q.QuoteTs > targetTime || validity > 0 && targetTime-q.QuoteTs > validity {
		return false
	}
	value := helpers.MustParseFloat(q.LastPrice)
	if kind == "FUNDING" {
		return !value.IsZero() || strings.TrimSpace(q.LastPrice) == "0"
	}
	return value.IsPositive()
}

func persistMarketSnapshot(ctx context.Context, model models.TTradeMarketSnapshotModel, tenantID, symbolID int64, s *cache.SettlementSnapshot) error {
	for field, value := range map[string]string{"price": s.Price, "mark_price": s.MarkPrice, "index_price": s.IndexPrice, "funding_rate": s.FundingRate} {
		if err := validateTradeDecimal(value); err != nil {
			return fmt.Errorf("authoritative snapshot %s exceeds Trade DECIMAL(36,18) contract: %w", field, err)
		}
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	confirmed := int64(common.YesNo_YES_NO_NO)
	if s.Confirmed {
		confirmed = int64(common.YesNo_YES_NO_YES)
	}
	_, err = model.InsertIgnore(ctx, &models.TTradeMarketSnapshot{TenantId: tenantID, SnapshotId: s.SnapshotID, SnapshotKind: s.Kind, SymbolId: symbolID, Source: s.Source, Price: helpers.MustParseFloat(s.Price), MarkPrice: helpers.MustParseFloat(s.MarkPrice), IndexPrice: helpers.MustParseFloat(s.IndexPrice), FundingRate: helpers.MustParseFloat(s.FundingRate), SourceTimestamp: s.SourceTimestamp, SnapshotTimestamp: s.SnapshotTimestamp, Revision: s.Revision, FormulaVersion: s.FormulaVersion, Confirmed: confirmed, RawPayload: string(raw), CreateTimes: utils.NowMillis()})
	return err
}

func validateTradeDecimal(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	d, err := decimal.NewFromString(value)
	if err != nil {
		return fmt.Errorf("invalid decimal %q", value)
	}
	plain := d.String()
	unsigned := strings.TrimPrefix(plain, "-")
	parts := strings.SplitN(unsigned, ".", 2)
	integerDigits := len(strings.TrimLeft(parts[0], "0"))
	if integerDigits == 0 {
		integerDigits = 1
	}
	fractionDigits := 0
	if len(parts) == 2 {
		fractionDigits = len(strings.TrimRight(parts[1], "0"))
	}
	if integerDigits > 18 || fractionDigits > 18 {
		return fmt.Errorf("value %q requires %d integer and %d fractional digits", value, integerDigits, fractionDigits)
	}
	return nil
}
func quoteIsValid(q *marketQuoteSnapshot, validity int64) bool {
	return quoteIsValidAt(q, utils.NowMillis(), validity)
}
func quoteIsValidAt(q *marketQuoteSnapshot, targetTime, validity int64) bool {
	return q != nil && q.Confirmed && q.SnapshotID != "" && helpers.MustParseFloat(q.LastPrice).IsPositive() && q.QuoteTs > 0 && q.QuoteTs <= targetTime && (validity <= 0 || targetTime-q.QuoteTs <= validity)
}
func parseQuoteSource(source string) (string, string, string) {
	parts := strings.Split(source, ":")
	if len(parts) >= 3 {
		return parts[0], parts[1], parts[2]
	}
	if len(parts) == 2 {
		return "", parts[0], parts[1]
	}
	return "", source, ""
}
func quoteSource(q *marketQuoteSnapshot) string { return q.Category + ":" + q.Market + ":" + q.Symbol }
func nonEmpty(v, f string) string {
	if v != "" {
		return v
	}
	return f
}
func secondsResult(direction int64, start, end, tolerance decimal.Decimal) trade.SecondsResult {
	delta := end.Sub(start)
	if delta.Abs().LessThanOrEqual(tolerance) {
		return trade.SecondsResult_SECONDS_RESULT_DRAW
	}
	up := delta.IsPositive()
	if direction == int64(trade.SecondsDirection_SECONDS_DIRECTION_UP) && up || direction == int64(trade.SecondsDirection_SECONDS_DIRECTION_DOWN) && !up {
		return trade.SecondsResult_SECONDS_RESULT_WIN
	}
	return trade.SecondsResult_SECONDS_RESULT_LOSE
}
func secondsPayout(stake, payoutRate, feeRate decimal.Decimal, result trade.SecondsResult) (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	if result != trade.SecondsResult_SECONDS_RESULT_WIN {
		return decimal.Zero, decimal.Zero, decimal.Zero
	}
	profit := contractmath.RoundCredit(stake.Mul(payoutRate))
	fee := contractmath.RoundDebit(profit.Mul(feeRate))
	return profit, fee, stake.Add(profit).Sub(fee)
}
func (l *ProcessSecondsSettlementsLogic) consumeSecondsStake(i *models.SecondsOrderWorkItem) error {
	r, e := l.svcCtx.AssetClient.DeductFrozenAssetByBizNo(l.ctx, &asset.DeductFrozenAssetByBizNoReq{TenantId: i.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: i.OrderNo, Amount: i.StakeAmount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: i.Id, BizNo: i.OrderNo + "-SECONDS-CONSUME", Remark: "seconds stake consume"})
	if e != nil {
		return e
	}
	return validateSecondsAssetResponse("consume", r)
}
func (l *ProcessSecondsSettlementsLogic) creditSeconds(i *models.SecondsOrderWorkItem, amount decimal.Decimal) error {
	r, e := l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{TenantId: i.TenantId, UserId: i.UserId, WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: i.StakeAsset, Amount: amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: i.Id, BizNo: i.OrderNo + "-SECONDS-PAYOUT", Remark: "seconds payout"})
	if e != nil {
		return e
	}
	return validateSecondsAssetResponse("payout", r)
}

type secondsAssetResponse interface {
	GetBase() *common.RespBase
}

func validateSecondsAssetResponse(action string, response secondsAssetResponse) error {
	if response == nil || response.GetBase() == nil {
		return fmt.Errorf("seconds %s returned an empty Asset response", action)
	}
	if response.GetBase().GetCode() != 200 {
		return fmt.Errorf("seconds %s rejected: %s", action, response.GetBase().GetMsg())
	}
	return nil
}
func insertSecondsPriceSnapshot(ctx context.Context, tx *models.TransactionModels, order *models.TTradeOrderSeconds, q *marketQuoteSnapshot, typ trade.SecondsPriceSnapshotType, selected bool) error {
	raw, _ := json.Marshal(q)
	yes := int64(common.YesNo_YES_NO_NO)
	if selected {
		yes = int64(common.YesNo_YES_NO_YES)
	}
	_, err := tx.TradeSecondsPriceSnapshot.Insert(ctx, &models.TTradeSecondsPriceSnapshot{TenantId: order.TenantId, OrderId: order.Id, SnapshotType: int64(typ), Source: quoteSource(q), Price: helpers.MustParseFloat(q.LastPrice), QuoteTime: q.QuoteTs, ReceivedAt: q.ReceivedAt, Algorithm: order.PriceAlgorithm, IsSelected: yes, RawPayload: sql.NullString{String: string(raw), Valid: true}, CreateTimes: utils.NowMillis()})
	return err
}
