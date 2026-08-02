package applogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const maxPublicOptionChainContracts = 500

type ListOptionChainLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOptionChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOptionChainLogic {
	return &ListOptionChainLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按标的和精确到期时间获取期权链、24小时成交统计及未平仓量
func (l *ListOptionChainLogic) ListOptionChain(in *option.ListOptionChainReq) (*option.ListOptionChainResp, error) {
	if in == nil || strings.TrimSpace(in.UnderlyingSymbol) == "" || in.ExpireTime <= 0 {
		return optionChainParamError(l.ctx), nil
	}
	status := in.Status
	if status == option.ContractStatus_CONTRACT_STATUS_UNKNOWN {
		status = option.ContractStatus_CONTRACT_STATUS_TRADING
	}
	if status != option.ContractStatus_CONTRACT_STATUS_TRADING &&
		status != option.ContractStatus_CONTRACT_STATUS_PAUSED {
		return optionChainParamError(l.ctx), nil
	}
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	windowStart := now - 24*60*60
	var contracts []*models.TOptionContract
	var markets []*models.TOptionMarket
	var trades []*models.OptionTradeStatistics
	var interests []*models.OptionOpenInterest
	err = withPublicMarketSnapshot(l.ctx, l.svcCtx.DB, func(ctx context.Context, conn sqlx.SqlConn) error {
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis)
		tradeModel := models.NewTOptionTradeModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)

		var queryErr error
		contracts, queryErr = contractModel.FindOptionChain(
			ctx, tenantID, in.UnderlyingSymbol, in.ExpireTime, int64(status),
			maxPublicOptionChainContracts+1,
		)
		if queryErr != nil {
			return queryErr
		}
		if len(contracts) > maxPublicOptionChainContracts {
			return errOptionChainTooBroad
		}
		if validateErr := validatePublicOptionChainContracts(contracts); validateErr != nil {
			return validateErr
		}
		contractIDs := make([]int64, 0, len(contracts))
		for _, contract := range contracts {
			contractIDs = append(contractIDs, contract.Id)
		}
		if markets, queryErr = marketModel.FindByContractIDs(ctx, tenantID, contractIDs); queryErr != nil {
			return queryErr
		}
		if trades, queryErr = tradeModel.FindStatisticsByContracts(ctx, tenantID, contractIDs, windowStart, now); queryErr != nil {
			return queryErr
		}
		if interests, queryErr = positionModel.FindOpenInterestByContracts(ctx, tenantID, contractIDs); queryErr != nil {
			return queryErr
		}
		return nil
	})
	if err != nil {
		if err == errOptionChainTooBroad {
			return optionChainParamError(l.ctx), nil
		}
		return nil, err
	}
	rows := buildOptionChainRows(contracts, markets, trades, interests, windowStart, now)
	return &option.ListOptionChainResp{
		Base:                  helper.OkResp(),
		Data:                  rows,
		GeneratedAt:           now,
		StatisticsWindowStart: windowStart,
	}, nil
}

var errOptionChainTooBroad = fmt.Errorf("option chain exceeds %d contracts", maxPublicOptionChainContracts)

func validatePublicOptionChainContracts(contracts []*models.TOptionContract) error {
	seen := make(map[string]int64, len(contracts))
	for _, contract := range contracts {
		key := fmt.Sprintf("%s/%d", conv.FloatString(contract.StrikePrice), contract.OptionType)
		if previousID, ok := seen[key]; ok {
			return fmt.Errorf(
				"duplicate public option chain leg: strike=%s type=%d contracts=%d,%d",
				conv.FloatString(contract.StrikePrice), contract.OptionType, previousID, contract.Id,
			)
		}
		seen[key] = contract.Id
	}
	return nil
}

func optionChainParamError(ctx context.Context) *option.ListOptionChainResp {
	return &option.ListOptionChainResp{
		Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx)),
	}
}

func buildOptionChainRows(
	contracts []*models.TOptionContract,
	markets []*models.TOptionMarket,
	trades []*models.OptionTradeStatistics,
	interests []*models.OptionOpenInterest,
	windowStart, now int64,
) []*option.OptionChainRow {
	marketByContract := make(map[int64]*models.TOptionMarket, len(markets))
	for _, market := range markets {
		marketByContract[market.ContractId] = market
	}
	tradeByContract := make(map[int64]*models.OptionTradeStatistics, len(trades))
	for _, statistic := range trades {
		tradeByContract[statistic.ContractId] = statistic
	}
	interestByContract := make(map[int64]*models.OptionOpenInterest, len(interests))
	for _, interest := range interests {
		interestByContract[interest.ContractId] = interest
	}
	rows := make([]*option.OptionChainRow, 0, (len(contracts)+1)/2)
	var row *option.OptionChainRow
	for _, contract := range contracts {
		strike := conv.FloatString(contract.StrikePrice)
		if row == nil || row.StrikePrice != strike {
			row = &option.OptionChainRow{StrikePrice: strike}
			rows = append(rows, row)
		}
		leg := &option.OptionChainLeg{
			Contract: helpers.ToContractProto(contract),
			Market:   helpers.ToMarketProto(marketByContract[contract.Id]),
			Statistics: buildPublicMarketStatistics(
				contract.Id, tradeByContract[contract.Id], interestByContract[contract.Id],
				windowStart, now,
			),
		}
		if contract.OptionType == int64(option.OptionType_OPTION_TYPE_CALL) {
			row.Call = leg
		} else if contract.OptionType == int64(option.OptionType_OPTION_TYPE_PUT) {
			row.Put = leg
		}
	}
	return rows
}

func buildPublicMarketStatistics(
	contractID int64,
	trade *models.OptionTradeStatistics,
	interest *models.OptionOpenInterest,
	windowStart, now int64,
) *option.OptionMarketStatistics {
	result := &option.OptionMarketStatistics{
		ContractId:            contractID,
		Volume_24H:            "0",
		Turnover_24H:          "0",
		OpenInterest:          "0",
		LongOpenInterest:      "0",
		ShortOpenInterest:     "0",
		OiBalanced:            true,
		StatisticsWindowStart: windowStart,
		StatisticsAsOf:        now,
		Source:                "OPTION_TRADES_AND_SETTLED_POSITIONS",
		OpenInterestMethod:    "MAX_LONG_SHORT",
	}
	if trade != nil {
		result.Volume_24H = conv.FloatString(trade.Volume)
		result.Turnover_24H = conv.FloatString(trade.Turnover)
		result.TradeCount_24H = trade.TradeCount
	}
	if interest != nil {
		result.LongOpenInterest = conv.FloatString(interest.LongQty)
		result.ShortOpenInterest = conv.FloatString(interest.ShortQty)
		result.OiBalanced = interest.LongQty.Equal(interest.ShortQty)
		result.PositionAsOf = interest.AsOf
		result.OpenInterest = conv.FloatString(decimal.Max(interest.LongQty, interest.ShortQty))
	}
	return result
}
