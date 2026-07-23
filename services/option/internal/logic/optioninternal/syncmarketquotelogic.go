package optioninternallogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
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

func NewSyncMarketQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncMarketQuoteLogic {
	return &SyncMarketQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步标的行情，更新对应期权合约行情和快照。
func (l *SyncMarketQuoteLogic) SyncMarketQuote(in *option.SyncMarketQuoteReq) (*option.InternalCommonResp, error) {
	if in == nil {
		return &option.InternalCommonResp{Base: helper.ErrResp(i18n.RequestRequired, i18n.Translate(i18n.RequestRequired, l.ctx))}, nil
	}
	symbol := strings.ToUpper(strings.TrimSpace(in.GetSymbol()))
	if symbol == "" {
		return &option.InternalCommonResp{Base: helper.ErrResp(i18n.SymbolRequired, i18n.Translate(i18n.SymbolRequired, l.ctx))}, nil
	}
	underlyingPrice, err := decimal.NewFromString(strings.TrimSpace(in.GetUnderlyingPrice()))
	if err != nil || !underlyingPrice.IsPositive() {
		return &option.InternalCommonResp{Base: helper.ErrResp(i18n.UnderlyingPriceMustBePositive, i18n.Translate(i18n.UnderlyingPriceMustBePositive, l.ctx))}, nil
	}

	now := time.Now().Unix()
	snapshotTime := normalizeQuoteTime(in.GetQuoteTs(), now)
	var cursor int64
	var updated int64

	for {
		contracts, _, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
			TenantId:         in.GetTenantId(),
			UnderlyingSymbol: symbol,
		}, cursor, 200)
		if err != nil {
			return nil, err
		}
		if len(contracts) == 0 {
			break
		}

		for _, contract := range contracts {
			cursor = contract.Id
			if !canSyncMarketQuote(contract.Status) {
				continue
			}
			if err := l.syncContractMarket(contract, underlyingPrice, snapshotTime, now); err != nil {
				return nil, err
			}
			updated++
		}
	}

	l.Infof("option market quote synced, symbol=%s market=%s category=%s updated=%d",
		symbol, in.GetMarket(), in.GetCategoryCode(), updated)
	return &option.InternalCommonResp{Base: helper.OkResp()}, nil
}

func (l *SyncMarketQuoteLogic) syncContractMarket(contract *models.TOptionContract, underlyingPrice decimal.Decimal, snapshotTime int64, now int64) error {
	intrinsicValue := calcIntrinsicValue(contract.OptionType, contract.StrikePrice, underlyingPrice)

	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
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

		return insertMarketSnapshot(ctx, snapshotModel, market, now)
	})
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
