package optionlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	marketEvent "wklive/common/market"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SyncMarketQuoteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type MarketSnapshotConsumeResult struct {
	Updated    int64
	Duplicates int64
}

func NewSyncMarketQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncMarketQuoteLogic {
	return &SyncMarketQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SyncMarketQuoteLogic) SyncAuthoritativeSnapshot(event marketEvent.AuthoritativeSnapshotEvent) (MarketSnapshotConsumeResult, error) {
	if event.Version != marketEvent.AuthoritativeSnapshotEventVersion {
		return MarketSnapshotConsumeResult{}, fmt.Errorf("unsupported authoritative snapshot event version: %d", event.Version)
	}
	if strings.TrimSpace(event.SnapshotID) == "" {
		return MarketSnapshotConsumeResult{}, errors.New("authoritative snapshot id is required")
	}
	if strings.TrimSpace(event.CategoryCode) == "" {
		return MarketSnapshotConsumeResult{}, errors.New("market category code is required")
	}
	if strings.TrimSpace(event.Market) == "" {
		return MarketSnapshotConsumeResult{}, errors.New("market is required")
	}
	result, err := l.syncMarketQuote(0, event)
	if err != nil {
		return MarketSnapshotConsumeResult{}, err
	}
	l.Infof("authoritative option market quote consumed, snapshotId=%s symbol=%s updated=%d duplicates=%d",
		event.SnapshotID, event.Symbol, result.Updated, result.Duplicates)
	return result, nil
}

func (l *SyncMarketQuoteLogic) syncMarketQuote(tenantID int64, event marketEvent.AuthoritativeSnapshotEvent) (MarketSnapshotConsumeResult, error) {
	symbol := strings.ToUpper(strings.TrimSpace(event.Symbol))
	if symbol == "" {
		return MarketSnapshotConsumeResult{}, errors.New("market symbol is required")
	}
	underlyingPrice, err := decimal.NewFromString(strings.TrimSpace(event.UnderlyingPrice))
	if err != nil || !underlyingPrice.IsPositive() {
		return MarketSnapshotConsumeResult{}, errors.New("underlying price must be positive")
	}
	now := time.Now().Unix()
	snapshotTime := normalizeQuoteTime(event.QuoteTimestamp, now)
	var cursor int64
	var result MarketSnapshotConsumeResult

	for {
		// Option contracts currently identify their underlying globally by symbol;
		// category and market are validated above but cannot be used as DB filters
		// until those dimensions are added to t_option_contract.
		contracts, _, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
			TenantId:         tenantID,
			UnderlyingSymbol: symbol,
		}, cursor, 200)
		if err != nil {
			return MarketSnapshotConsumeResult{}, err
		}
		if len(contracts) == 0 {
			break
		}

		for _, contract := range contracts {
			cursor = contract.Id
			if !canSyncMarketQuote(contract.Status) {
				continue
			}
			changed, err := l.syncContractMarket(contract, event.SnapshotID, underlyingPrice, snapshotTime, now)
			if err != nil {
				return MarketSnapshotConsumeResult{}, err
			}
			if changed {
				result.Updated++
			} else if event.SnapshotID != "" {
				result.Duplicates++
			}
		}
	}
	return result, nil
}

func (l *SyncMarketQuoteLogic) syncContractMarket(contract *models.TOptionContract, snapshotID string, underlyingPrice decimal.Decimal, snapshotTime int64, now int64) (bool, error) {
	intrinsicValue := calcIntrinsicValue(contract.OptionType, contract.StrikePrice, underlyingPrice)
	changed := false
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		if snapshotID != "" {
			inboxModel := models.NewTOptionMarketSnapshotInboxModel(conn, l.svcCtx.Config.CacheRedis)
			claimed, err := inboxModel.Claim(ctx, snapshotID, contract.TenantId, contract.Id, now)
			if err != nil {
				return err
			}
			if !claimed {
				return nil
			}
		}
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis)
		snapshotModel := models.NewTOptionMarketSnapshotModel(conn, l.svcCtx.Config.CacheRedis)

		market, err := marketModel.FindOneByTenantIdContractId(ctx, contract.TenantId, contract.Id)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if market == nil {
			market = &models.TOptionMarket{
				TenantId:    contract.TenantId,
				ContractId:  contract.Id,
				CreateTimes: now,
			}
		}

		market.UnderlyingPrice = underlyingPrice
		market.IntrinsicValue = intrinsicValue
		if market.MarkPrice.IsPositive() {
			market.TimeValue = decimal.Max(market.MarkPrice.Sub(intrinsicValue), decimal.Zero)
		}
		market.SnapshotTime = snapshotTime
		market.UpdateTimes = now

		if market.Id == 0 {
			result, err := marketModel.Insert(ctx, market)
			if err != nil {
				return err
			}
			market.Id, _ = result.LastInsertId()
		} else if err := marketModel.Update(ctx, market); err != nil {
			return err
		}

		if err := insertMarketSnapshot(ctx, snapshotModel, market, now); err != nil {
			return err
		}
		changed = true
		return nil
	})
	return changed, err
}

func canSyncMarketQuote(status int64) bool {
	switch option.ContractStatus(status) {
	case option.ContractStatus_CONTRACT_STATUS_PENDING,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		option.ContractStatus_CONTRACT_STATUS_PAUSED,
		option.ContractStatus_CONTRACT_STATUS_EXPIRED:
		return true
	default:
		return false
	}
}

func normalizeQuoteTime(ts int64, fallback int64) int64 {
	if ts <= 0 {
		return fallback
	}
	if ts > 1_000_000_000_000 {
		return ts / 1000
	}
	return ts
}

func calcIntrinsicValue(optionType int64, strikePrice decimal.Decimal, underlyingPrice decimal.Decimal) decimal.Decimal {
	switch option.OptionType(optionType) {
	case option.OptionType_OPTION_TYPE_CALL:
		return decimal.Max(underlyingPrice.Sub(strikePrice), decimal.Zero)
	case option.OptionType_OPTION_TYPE_PUT:
		return decimal.Max(strikePrice.Sub(underlyingPrice), decimal.Zero)
	default:
		return decimal.Zero
	}
}
