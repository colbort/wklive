package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"wklive/services/trade/internal/logic/helpers"

	cache "wklive/common/market"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/domain/contractmath"
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
			if symbol.ContractType != int64(common.ContractType_CONTRACT_TYPE_PERPETUAL) {
				continue
			}
			interval := c.FundingIntervalMinutes * 60 * 1000
			currentSettlementTime := now / interval * interval
			if currentSettlementTime <= 0 {
				continue
			}
			settlementTime := currentSettlementTime
			recoverHistory := true
			latest, latestErr := l.svcCtx.ContractFundingBatchModel.FindLatestBySymbol(l.ctx, c.TenantId, c.SymbolId)
			if latestErr == nil {
				if latest.Status != int64(trade.FundingBatchStatus_FUNDING_BATCH_STATUS_COMPLETED) {
					continue
				}
				settlementTime = latest.SettlementTime + interval
				if settlementTime > currentSettlementTime {
					continue
				}
				recoverHistory = true
			} else if !errors.Is(latestErr, models.ErrNotFound) {
				return latestErr
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
			fundingFact, err := l.svcCtx.TradeMarketSnapshotModel.FindOneByTenantIdSymbolIdSnapshotId(l.ctx, c.TenantId, c.SymbolId, source)
			if err != nil {
				return fmt.Errorf("reload immutable funding input fact: %w", err)
			}
			if fundingFact.FormulaVersion == "" {
				return errors.New("immutable funding input fact has no formula version")
			}
			batchNo := fmt.Sprintf("FND-%d-%d", c.SymbolId, settlementTime)
			if err := helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
				conn := sqlx.NewSqlConnFromSession(session)
				bm := models.NewTContractFundingBatchModel(conn, l.svcCtx.Config.CacheRedis)
				sm := models.NewTContractFundingSettlementModel(conn, l.svcCtx.Config.CacheRedis)
				pm := models.NewTContractPositionModel(conn, l.svcCtx.Config.CacheRedis)
				var active []*models.TContractPosition
				if recoverHistory {
					history, historyErr := models.NewTContractPositionHistoryModel(conn, l.svcCtx.Config.CacheRedis).FindLatestBySymbolAt(ctx, c.TenantId, c.SymbolId, settlementTime)
					if historyErr != nil {
						return historyErr
					}
					active = fundingPositionsFromHistory(history)
				} else {
					active, err = pm.FindActiveListForUpdate(ctx, c.TenantId, c.SymbolId)
					if err != nil {
						return err
					}
				}
				res, err := bm.Insert(ctx, &models.TContractFundingBatch{TenantId: c.TenantId, BatchNo: batchNo, SymbolId: c.SymbolId, FundingRate: rate, MarkPrice: mark, IndexPrice: index, PriceSource: source, FormulaVersion: fundingFact.FormulaVersion, SettlementTime: settlementTime, Status: int64(trade.FundingBatchStatus_FUNDING_BATCH_STATUS_SETTLING), TotalPositions: int64(len(active)), CreateTimes: now, UpdateTimes: now})
				if err != nil {
					return err
				}
				batchID, _ := res.LastInsertId()
				feeTotals := make(map[string]decimal.Decimal)
				im := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
				for _, p := range active {
					if p.MarginAsset == "" {
						return fmt.Errorf("funding position %d has no immutable margin asset", p.Id)
					}
					values, err := contractmath.CalculateTradeValues(p.ContractValueType, p.Qty, c.ContractSize, mark)
					if err != nil {
						return err
					}
					fee := contractmath.RoundCredit(values.SettlementNotional.Mul(rate))
					if p.PositionSide == int64(trade.PositionSide_POSITION_SIDE_LONG) {
						fee = fee.Neg()
					}
					feeTotals[p.MarginAsset] = feeTotals[p.MarginAsset].Add(fee)
					settlementNo := fmt.Sprintf("%s-%d", batchNo, p.Id)
					if _, err = sm.Insert(ctx, &models.TContractFundingSettlement{TenantId: c.TenantId, SettlementNo: settlementNo, BatchId: batchID, BatchNo: batchNo, SymbolId: c.SymbolId, UserId: p.UserId, PositionId: p.Id, PositionSide: p.PositionSide, FundingRate: rate, MarkPrice: mark, PositionQty: p.Qty, PositionVersion: p.Version, FeeAsset: p.MarginAsset, FeeAmount: fee, SettlementTime: settlementTime, Status: int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now}); err != nil {
						return err
					}
					if !fee.IsZero() {
						action, step := trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE, int64(2)
						if fee.IsNegative() {
							action, step = trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS, 1
						}
						if err = insertSettlementInstructionIdempotent(ctx, im, &models.TTradeSettlementInstruction{TenantId: c.TenantId, InstructionNo: settlementNo + "-ASSET", BizType: "funding", BizId: settlementNo, BatchNo: batchNo, PositionId: p.Id, UserId: p.UserId, Action: int64(action), Asset: p.MarginAsset, Amount: fee.Abs(), StepNo: step, Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now}); err != nil {
							return err
						}
					}
				}
				for assetCode, difference := range feeTotals {
					if difference.IsZero() {
						continue
					}
					account, accountErr := l.svcCtx.FundingDifferenceAcctModel.FindEnabled(ctx, c.TenantId, assetCode)
					if accountErr != nil {
						return fmt.Errorf("funding difference account unavailable: asset=%s difference=%s: %w", assetCode, difference, accountErr)
					}
					action, step := fundingDifferenceInstruction(difference)
					instructionNo := fmt.Sprintf("%s-DIFF-%s", batchNo, assetCode)
					if err = insertSettlementInstructionIdempotent(ctx, im, &models.TTradeSettlementInstruction{TenantId: c.TenantId, InstructionNo: instructionNo, BizType: "funding", BizId: "DIFF:" + assetCode, BatchNo: batchNo, UserId: account.FundUserId, Action: int64(action), Asset: assetCode, Amount: difference.Abs(), StepNo: step, Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now}); err != nil {
						return err
					}
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

func fundingPositionsFromHistory(history []*models.TContractPositionHistory) []*models.TContractPosition {
	positions := make([]*models.TContractPosition, 0, len(history))
	for _, h := range history {
		if h == nil || !h.AfterQty.IsPositive() {
			continue
		}
		positions = append(positions, &models.TContractPosition{Id: h.PositionId, TenantId: h.TenantId, UserId: h.UserId, SymbolId: h.SymbolId, ContractType: h.ContractType, ContractValueType: h.ContractValueType, PositionSide: h.PositionSide, Qty: h.AfterQty, MarginAsset: h.MarginAsset, Version: h.AfterVersion})
	}
	return positions
}

func fundingDifferenceInstruction(userNet decimal.Decimal) (trade.SettlementInstructionAction, int64) {
	if userNet.IsNegative() {
		return trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE, 2
	}
	return trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS, 1
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
	fundingQ, _, err := quotes.getValidQuotesAtKind("FUNDING_RATE", c.MarkPriceSource, c.SymbolId, settlementTime, 30_000)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", err
	}
	mark, index := helpers.MustParseFloat(markQ.LastPrice), helpers.MustParseFloat(indexQ.LastPrice)
	if !index.IsPositive() {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", errors.New("invalid funding index price")
	}
	rate := helpers.MustParseFloat(fundingQ.LastPrice)
	if c.FundingRateCap.IsPositive() && rate.GreaterThan(c.FundingRateCap) {
		rate = c.FundingRateCap
	}
	if c.FundingRateFloor.IsNegative() && rate.LessThan(c.FundingRateFloor) {
		rate = c.FundingRateFloor
	}
	category, market, symbol := parseQuoteSource(c.MarkPriceSource)
	snapshot := &cache.SettlementSnapshot{Kind: "FUNDING", CategoryCode: category, Market: market, Symbol: symbol, MarkPrice: mark.String(), IndexPrice: index.String(), FundingRate: rate.String(), Source: fundingInputIdentity(markQ, indexQ, fundingQ), SourceTimestamp: minInt64(minInt64(markQ.QuoteTs, indexQ.QuoteTs), fundingQ.QuoteTs), SnapshotTimestamp: utils.NowMillis(), Revision: maxInt64(maxInt64(markQ.Revision, indexQ.Revision), fundingQ.Revision), FormulaVersion: nonEmpty(fundingQ.FormulaVersion, "price-engine"), Authority: strings.TrimSpace(l.svcCtx.Config.PriceEngineAuthority), Confirmed: markQ.Confirmed && indexQ.Confirmed && fundingQ.Confirmed}
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
func fundingInputIdentity(mark, index, funding *marketQuoteSnapshot) string {
	return fmt.Sprintf("MARK=%s|INDEX=%s|FUNDING=%s", mark.SnapshotID, index.SnapshotID, funding.SnapshotID)
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
	// Zero-value settlements have no Asset step but still need the position projection.
	rows, _, err := l.svcCtx.ContractFundingSettleModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: tenantID, Status: int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_PENDING)}, 0, 100)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.FeeAmount.IsZero() {
			if err = l.completeFunding(row); err != nil {
				return err
			}
		}
	}
	for processed := 0; processed < 1000; {
		now := utils.NowMillis()
		instructions, findErr := l.svcCtx.TradeSettlementInstrModel.FindPendingBiz(l.ctx, tenantID, "funding", now, 100)
		if findErr != nil {
			return findErr
		}
		if len(instructions) == 0 {
			break
		}
		progressed := false
		for _, instruction := range instructions {
			claimed, lease, claimErr := l.svcCtx.TradeSettlementInstrModel.ClaimLease(l.ctx, instruction.Id, now)
			if claimErr != nil {
				return claimErr
			}
			if !claimed {
				continue
			}
			instruction.UpdateTimes = lease
			progressed = true
			processed++
			if executeErr := l.executeFundingInstruction(instruction); executeErr != nil {
				if markErr := l.failFundingInstruction(instruction, executeErr); markErr != nil {
					return markErr
				}
			}
		}
		if !progressed {
			break
		}
	}
	return l.finishBatches(tenantID)
}

func (l *ProcessFundingSettlementsLogic) executeFundingInstruction(item *models.TTradeSettlementInstruction) error {
	walletType := common.WalletType_WALLET_TYPE_CONTRACT
	if item.PositionId == 0 {
		account, err := l.svcCtx.FundingDifferenceAcctModel.FindEnabled(l.ctx, item.TenantId, item.Asset)
		if err != nil {
			return fmt.Errorf("resolve funding difference account: %w", err)
		}
		if account == nil || account.FundUserId != item.UserId || account.WalletType <= 0 {
			return errors.New("funding difference account identity changed")
		}
		walletType = common.WalletType(account.WalletType)
	} else if err := l.validateFundingUserInstruction(item); err != nil {
		return err
	}
	var resp *asset.ChangeAssetResp
	var err error
	requestID := item.Id
	if item.Action == int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE) {
		resp, err = l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: walletType, Coin: item.Asset, Amount: item.Amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: requestID, BizNo: item.InstructionNo, Remark: "contract funding saga credit"})
	} else if item.Action == int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS) {
		resp, err = l.svcCtx.AssetClient.SubAvailable(l.ctx, &asset.SubAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: walletType, Coin: item.Asset, Amount: item.Amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: requestID, BizNo: item.InstructionNo, Remark: "contract funding saga debit"})
	} else {
		return fmt.Errorf("invalid funding instruction action: %d", item.Action)
	}
	if err != nil {
		return err
	}
	if resp == nil || resp.GetBase() == nil {
		return errors.New("funding asset instruction returned an empty response")
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("funding asset instruction rejected: code=%d msg=%s", resp.GetBase().GetCode(), resp.GetBase().GetMsg())
	}
	return l.completeFundingInstruction(item)
}

func (l *ProcessFundingSettlementsLogic) validateFundingUserInstruction(item *models.TTradeSettlementInstruction) error {
	settlement, err := l.svcCtx.ContractFundingSettleModel.FindOneByTenantIdSettlementNo(l.ctx, item.TenantId, item.BizId)
	if err != nil {
		return err
	}
	expectedAction := int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE)
	if settlement.FeeAmount.IsNegative() {
		expectedAction = int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS)
	}
	if settlement.BatchNo != item.BatchNo || settlement.PositionId != item.PositionId || settlement.UserId != item.UserId || settlement.FeeAsset != item.Asset || !settlement.FeeAmount.Abs().Equal(item.Amount) || item.Action != expectedAction || settlement.FeeAmount.IsZero() {
		return errors.New("funding instruction does not match immutable settlement facts")
	}
	position, err := l.svcCtx.ContractPositionModel.FindOne(l.ctx, item.PositionId)
	if err != nil {
		return err
	}
	if position.TenantId != item.TenantId || position.UserId != item.UserId || position.LastFundingTime >= settlement.SettlementTime {
		return errors.New("funding instruction position identity or settlement time is invalid")
	}
	return nil
}

func (l *ProcessFundingSettlementsLogic) completeFundingInstruction(item *models.TTradeSettlementInstruction) error {
	now := utils.NowMillis()
	return helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		im := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := im.FindOneForUpdate(ctx, item.Id)
		if err != nil || current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			return err
		}
		if !settlementInstructionLeaseOwned(current, item) {
			return errors.New("funding instruction lease lost")
		}
		if current.PositionId > 0 {
			settlement, findErr := models.NewTContractFundingSettlementModel(conn, l.svcCtx.Config.CacheRedis).FindOneByTenantIdSettlementNo(ctx, current.TenantId, current.BizId)
			if findErr != nil {
				return findErr
			}
			if err = l.completeFundingInSession(ctx, conn, settlement, now); err != nil {
				return err
			}
		}
		current.Status, current.NextRetryAt, current.LastErrorMsg, current.UpdateTimes = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS), 0, "", now
		return im.Update(ctx, current)
	})
}

func (l *ProcessFundingSettlementsLogic) failFundingInstruction(item *models.TTradeSettlementInstruction, cause error) error {
	now := utils.NowMillis()
	return helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
		im := models.NewTTradeSettlementInstructionModel(sqlx.NewSqlConnFromSession(session), l.svcCtx.Config.CacheRedis)
		current, err := im.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if !settlementInstructionLeaseOwned(current, item) {
			return nil
		}
		current.RetryCount++
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED)
		current.NextRetryAt = now + helpers.TradeEventRetryDelay(current.RetryCount).Milliseconds()
		if current.RetryCount >= 20 {
			current.Status, current.NextRetryAt = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW), 0
		}
		current.LastErrorMsg, current.UpdateTimes = cause.Error(), now
		if err = im.Update(ctx, current); err != nil {
			return err
		}
		if current.PositionId > 0 {
			sm := models.NewTContractFundingSettlementModel(sqlx.NewSqlConnFromSession(session), l.svcCtx.Config.CacheRedis)
			settlement, findErr := sm.FindOneByTenantIdSettlementNo(ctx, current.TenantId, current.BizId)
			if findErr != nil {
				return findErr
			}
			settlement.RetryCount = current.RetryCount
			settlement.NextRetryAt = current.NextRetryAt
			settlement.LastErrorMsg = current.LastErrorMsg
			settlement.UpdateTimes = now
			settlement.Status = int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_FAILED)
			if current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW) {
				settlement.Status = int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_MANUAL_REVIEW)
			}
			return sm.Update(ctx, settlement)
		}
		return nil
	})
}

func (l *ProcessFundingSettlementsLogic) completeFunding(row *models.TContractFundingSettlement) error {
	now := utils.NowMillis()
	return helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		return l.completeFundingInSession(ctx, conn, row, now)
	})
}

func (l *ProcessFundingSettlementsLogic) completeFundingInSession(ctx context.Context, conn sqlx.SqlConn, row *models.TContractFundingSettlement, now int64) error {
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
		if err = writeSystemPositionHistory(ctx, hm, before, current, row.SettlementTime, row.SettlementNo, trade.PositionActionType_POSITION_ACTION_TYPE_FUNDING_FEE, row.FeeAmount, decimal.Zero, row.MarkPrice, "funding fee settlement"); err != nil {
			return err
		}
	}
	row.Status = int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_SETTLED)
	row.SettledAt, row.NextRetryAt, row.LastErrorMsg, row.UpdateTimes = now, 0, "", now
	return sm.Update(ctx, row)
}
func (l *ProcessFundingSettlementsLogic) finishBatches(tenantID int64) error {
	batches, _, err := l.svcCtx.ContractFundingBatchModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: tenantID, Status: int64(trade.FundingBatchStatus_FUNDING_BATCH_STATUS_SETTLING)}, 0, 100)
	if err != nil {
		return err
	}
	for _, b := range batches {
		manual, manualErr := l.svcCtx.TradeSettlementInstrModel.CountByBatchStatus(l.ctx, b.TenantId, "funding", b.BatchNo, int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW))
		if manualErr != nil {
			return manualErr
		}
		unfinished, instructionErr := l.svcCtx.TradeSettlementInstrModel.CountUnfinishedByBatch(l.ctx, b.TenantId, "funding", b.BatchNo)
		if instructionErr != nil {
			return instructionErr
		}
		unreconciled, reconcileErr := l.svcCtx.TradeSettlementInstrModel.CountUnreconciledByBatch(l.ctx, b.TenantId, "funding", b.BatchNo)
		if reconcileErr != nil {
			return reconcileErr
		}
		count, err := l.svcCtx.ContractFundingSettleModel.CountByBatchStatus(l.ctx, b.TenantId, b.Id, int64(trade.FundingSettlementStatus_FUNDING_SETTLEMENT_STATUS_SETTLED))
		if err != nil {
			return err
		}
		b.SettledPositions = count
		if manual > 0 {
			b.Status = int64(trade.FundingBatchStatus_FUNDING_BATCH_STATUS_MANUAL_REVIEW)
			b.LastErrorMsg = "one or more funding asset instructions require manual review"
		} else if count == b.TotalPositions && unfinished == 0 && unreconciled == 0 {
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
