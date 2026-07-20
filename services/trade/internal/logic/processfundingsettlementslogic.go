package logic

import (
	"context"
	"errors"
	"fmt"

	marketcache "wklive/common/market"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessFundingSettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProcessFundingSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessFundingSettlementsLogic {
	return &ProcessFundingSettlementsLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ProcessFundingSettlementsLogic) Process(tenantID int64) error {
	if err := l.createDueBatches(tenantID); err != nil {
		return err
	}
	return l.settlePending(tenantID)
}

func (l *ProcessFundingSettlementsLogic) createDueBatches(tenantID int64) error {
	now := utils.NowMillis()
	cursor := int64(0)
	for {
		contracts, _, err := l.svcCtx.TradeSymbolContractModel.FindPage(l.ctx, cursor, 100)
		if err != nil {
			return err
		}
		if len(contracts) == 0 {
			return nil
		}
		for _, c := range contracts {
			cursor = c.Id
			if tenantID > 0 && c.TenantId != tenantID || c.FundingIntervalMinutes <= 0 {
				continue
			}
			symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, c.SymbolId)
			if err != nil {
				return err
			}
			if symbol.ContractType != int64(trade.ContractType_CONTRACT_TYPE_PERPETUAL) {
				continue
			}
			interval := c.FundingIntervalMinutes * 60 * 1000
			settlementTime := now / interval * interval
			if settlementTime <= 0 {
				continue
			}
			if _, err = l.svcCtx.ContractFundingBatchModel.FindOneByTenantIdSymbolIdSettlementTime(l.ctx, c.TenantId, c.SymbolId, settlementTime); err == nil {
				continue
			} else if !errors.Is(err, models.ErrNotFound) {
				return err
			}
			mark, index, rate, source, err := l.lockFundingInputs(c, settlementTime)
			if err != nil {
				return err
			}
			batchNo := fmt.Sprintf("FND-%d-%d", c.SymbolId, settlementTime)
			if err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				conn := sqlx.NewSqlConnFromSession(session)
				bm := models.NewTContractFundingBatchModel(conn, l.svcCtx.Config.CacheRedis)
				sm := models.NewTContractFundingSettlementModel(conn, l.svcCtx.Config.CacheRedis)
				pm := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)
				active, err := pm.FindActiveListForUpdate(ctx, c.TenantId, c.SymbolId)
				if err != nil {
					return err
				}
				res, err := bm.Insert(ctx, &models.TContractFundingBatch{TenantId: c.TenantId, BatchNo: batchNo, SymbolId: c.SymbolId, FundingRate: rate, MarkPrice: mark, IndexPrice: index, PriceSource: source, FormulaVersion: "premium-v1", SettlementTime: settlementTime, Status: int64(trade.FundingBatchStatus_FUNDING_BATCH_STATUS_SETTLING), TotalPositions: int64(len(active)), CreateTimes: now, UpdateTimes: now})
				if err != nil {
					return err
				}
				batchID, _ := res.LastInsertId()
				feeTotals := make(map[string]decimal.Decimal)
				for _, p := range active {
					values, err := calculateContractTradeValues(p.ContractValueType, p.Qty, c.ContractSize, mark)
					if err != nil {
						return err
					}
					fee := roundContractCredit(values.SettlementNotional.Mul(rate))
					if p.PositionSide == int64(trade.PositionSide_POSITION_SIDE_LONG) {
						fee = fee.Neg()
					}
					feeTotals[p.MarginAsset] = feeTotals[p.MarginAsset].Add(fee)
					if _, err = sm.Insert(ctx, &models.TContractFundingSettlement{TenantId: c.TenantId, SettlementNo: fmt.Sprintf("%s-%d", batchNo, p.Id), BatchId: batchID, BatchNo: batchNo, SymbolId: c.SymbolId, UserId: p.UserId, PositionId: p.Id, PositionSide: p.PositionSide, FundingRate: rate, MarkPrice: mark, PositionQty: p.Qty, PositionVersion: p.Version, FeeAsset: p.MarginAsset, FeeAmount: fee, SettlementTime: settlementTime, Status: int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now}); err != nil {
						return err
					}
				}
				if err = validateFundingConservation(feeTotals); err != nil {
					return err
				}
				return nil
			}); err != nil {
				return err
			}
		}
		if len(contracts) < 100 {
			return nil
		}
	}
}

func validateFundingConservation(totals map[string]decimal.Decimal) error {
	for asset, total := range totals {
		if !total.IsZero() {
			return fmt.Errorf("funding batch is not balanced: asset=%s difference=%s", asset, total)
		}
	}
	return nil
}

func (l *ProcessFundingSettlementsLogic) lockFundingInputs(c *models.TTradeSymbolContract, settlementTime int64) (decimal.Decimal, decimal.Decimal, decimal.Decimal, string, error) {
	if c.FundingRateSource != "premium-v1" {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", fmt.Errorf("unsupported authoritative funding rate source: %q", c.FundingRateSource)
	}
	quotes := NewProcessSecondsSettlementsLogic(l.ctx, l.svcCtx)
	markQ, _, err := quotes.getValidQuotesAtKind("MARK_PRICE", c.MarkPriceSource, c.SymbolId, settlementTime, 30_000)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", err
	}
	category, market, _ := parseQuoteSource(c.MarkPriceSource)
	indexSource := c.IndexSymbol
	if category != "" && market != "" && len(parseParts(indexSource)) < 3 {
		indexSource = category + ":" + market + ":" + indexSource
	}
	indexQ, _, err := quotes.getValidQuotesAtKind("INDEX_PRICE", indexSource, c.SymbolId, settlementTime, 30_000)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", err
	}
	mark, index := mustParseFloat(markQ.LastPrice), mustParseFloat(indexQ.LastPrice)
	if !index.IsPositive() {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", errors.New("invalid funding index price")
	}
	rate := mark.Sub(index).Div(index)
	if c.FundingRateCap.IsPositive() && rate.GreaterThan(c.FundingRateCap) {
		rate = c.FundingRateCap
	}
	if c.FundingRateFloor.IsNegative() && rate.LessThan(c.FundingRateFloor) {
		rate = c.FundingRateFloor
	}
	snapshot := &marketcache.SettlementSnapshot{Kind: "FUNDING", MarkPrice: mark.String(), IndexPrice: index.String(), FundingRate: rate.String(), Source: quoteSource(markQ) + "|" + quoteSource(indexQ), SourceTimestamp: minInt64(markQ.QuoteTs, indexQ.QuoteTs), SnapshotTimestamp: utils.NowMillis(), Revision: maxInt64(markQ.Revision, indexQ.Revision), FormulaVersion: "premium-v1", Confirmed: markQ.Confirmed && indexQ.Confirmed}
	if err := l.svcCtx.MarketDataCache.PutSettlementSnapshot(l.ctx, snapshot); err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", err
	}
	// Read-after-write proves that the exact immutable snapshot is available to audit/replay.
	if _, err := l.svcCtx.MarketDataCache.GetSettlementSnapshot(l.ctx, snapshot.SnapshotID); err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", err
	}
	if err := persistMarketSnapshot(l.ctx, l.svcCtx.TradeMarketSnapshotModel, c.TenantId, c.SymbolId, snapshot); err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", err
	}
	return mark, index, rate, snapshot.SnapshotID, nil
}
func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
func parseParts(v string) []string {
	_, m, s := parseQuoteSource(v)
	r := []string{}
	if m != "" {
		r = append(r, m)
	}
	if s != "" {
		r = append(r, s)
	}
	return r
}

func (l *ProcessFundingSettlementsLogic) settlePending(tenantID int64) error {
	for _, status := range []trade.FundingSettlementStatus{trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_PENDING, trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_FAILED} {
		cursor := int64(0)
		for {
			rows, _, err := l.svcCtx.ContractFundingSettleModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: tenantID, Status: int64(status)}, cursor, 100)
			if err != nil {
				return err
			}
			if len(rows) == 0 {
				break
			}
			for _, row := range rows {
				cursor = row.Id
				if row.NextRetryAt > utils.NowMillis() {
					continue
				}
				if row.FeeAmount.IsPositive() {
					ready, readyErr := l.fundingReceiversReady(row)
					if readyErr != nil {
						return readyErr
					}
					if !ready {
						continue
					}
				}
				if err := l.settleOne(row); err != nil {
					row.RetryCount++
					row.LastErrorMsg = err.Error()
					row.NextRetryAt = utils.NowMillis() + tradeEventRetryDelay(row.RetryCount).Milliseconds()
					row.Status = int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_FAILED)
					row.UpdateTimes = utils.NowMillis()
					if row.RetryCount >= 20 {
						row.Status = int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_MANUAL_REVIEW)
					}
					if updateErr := l.svcCtx.ContractFundingSettleModel.Update(l.ctx, row); updateErr != nil {
						return updateErr
					}
				}
			}
			if len(rows) < 100 {
				break
			}
		}
	}
	return l.finishBatches(tenantID)
}

func (l *ProcessFundingSettlementsLogic) fundingReceiversReady(row *models.TContractFundingSettlement) (bool, error) {
	count, err := l.svcCtx.ContractFundingSettleModel.CountUnsettledPayers(l.ctx, row.TenantId, row.BatchId, int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_SETTLED))
	return count == 0, err
}
func (l *ProcessFundingSettlementsLogic) settleOne(row *models.TContractFundingSettlement) error {
	if row.FeeAmount.IsZero() {
		return l.completeFunding(row)
	}
	var resp *asset.ChangeAssetResp
	var err error
	amount := row.FeeAmount.Abs().String()
	if row.FeeAmount.IsPositive() {
		resp, err = l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{TenantId: row.TenantId, UserId: row.UserId, WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: row.FeeAsset, Amount: amount, BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: row.Id, BizNo: row.SettlementNo, Remark: "contract funding income"})
	} else {
		resp, err = l.svcCtx.AssetClient.SubAvailable(l.ctx, &asset.SubAvailableReq{TenantId: row.TenantId, UserId: row.UserId, WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: row.FeeAsset, Amount: amount, BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: row.Id, BizNo: row.SettlementNo, Remark: "contract funding payment"})
	}
	if err != nil {
		return err
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("funding asset rejected: %s", resp.GetBase().GetMsg())
	}
	return l.completeFunding(row)
}
func (l *ProcessFundingSettlementsLogic) completeFunding(row *models.TContractFundingSettlement) error {
	now := utils.NowMillis()
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		sm := models.NewTContractFundingSettlementModel(conn, l.svcCtx.Config.CacheRedis)
		pm := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)
		hm := models.NewTContractPositionHistoryModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := pm.FindOneForUpdate(ctx, row.PositionId)
		if err != nil {
			return err
		}
		if current.Id != row.PositionId {
			return errors.New("funding settlement position identity changed")
		}
		if current.LastFundingTime < row.SettlementTime {
			before := cloneContractPosition(current)
			current.LastFundingTime = row.SettlementTime
			current.RealizedPnl = current.RealizedPnl.Add(row.FeeAmount)
			current.Version++
			current.UpdateTimes = now
			if err = pm.Update(ctx, current); err != nil {
				return err
			}
			if err = writeSystemPositionHistory(ctx, hm, before, current, row.SettlementNo, trade.PositionActionType_POSITION_ACTION_TYPE_FUNDING_FEE, row.FeeAmount, decimal.Zero, row.MarkPrice, "funding fee settlement"); err != nil {
				return err
			}
		}
		row.Status = int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_SETTLED)
		row.SettledAt, row.NextRetryAt, row.LastErrorMsg, row.UpdateTimes = now, 0, "", now
		return sm.Update(ctx, row)
	})
}
func (l *ProcessFundingSettlementsLogic) finishBatches(tenantID int64) error {
	batches, _, err := l.svcCtx.ContractFundingBatchModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: tenantID, Status: int64(trade.FundingBatchStatus_FUNDING_BATCH_STATUS_SETTLING)}, 0, 100)
	if err != nil {
		return err
	}
	for _, b := range batches {
		count, err := l.svcCtx.ContractFundingSettleModel.CountByBatchStatus(l.ctx, b.TenantId, b.Id, int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_SETTLED))
		if err != nil {
			return err
		}
		b.SettledPositions = count
		if count == b.TotalPositions {
			b.Status = int64(trade.FundingBatchStatus_FUNDING_BATCH_STATUS_COMPLETED)
		}
		b.Version++
		b.UpdateTimes = utils.NowMillis()
		if err := l.svcCtx.ContractFundingBatchModel.Update(l.ctx, b); err != nil {
			return err
		}
	}
	return nil
}
