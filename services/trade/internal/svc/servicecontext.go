package svc

import (
	"context"
	"fmt"
	"os"
	"time"
	bus "wklive/common/bus/redis"
	cache "wklive/common/market"
	"wklive/services/trade/internal/config"
	"wklive/services/trade/models"

	"wklive/proto/asset"

	v9 "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                      config.Config
	DB                          sqlx.SqlConn
	Redis                       *redis.Redis
	TaskSubscriber              *bus.Subscriber
	TradeEventPublisher         *bus.Publisher
	TradeEventSubscriber        *bus.Subscriber
	TradeEventInstanceID        string
	TradeSymbolModel            models.TTradeSymbolModel
	TradeSymbolSpotModel        models.TTradeSymbolSpotModel
	TradeSymbolContractModel    models.TTradeSymbolContractModel
	TradeUserConfigModel        models.TTradeUserConfigModel
	TradeOrderModel             models.TTradeOrderModel
	TradeOrderSpotModel         models.TTradeOrderSpotModel
	TradeOrderContractModel     models.TTradeOrderContractModel
	TradeOrderSecondsModel      models.TTradeOrderSecondsModel
	TradeSecondsPriceModel      models.TTradeSecondsPriceSnapshotModel
	TradeFillModel              models.TTradeFillModel
	TradeCancelLogModel         models.TTradeCancelLogModel
	ContractPositionModel       models.TContractPositionModel
	ContractPositionHistModel   models.TContractPositionHistoryModel
	ContractMarginSnapshotModel models.TContractMarginSnapshotModel
	ContractRiskLimitTierModel  models.TContractRiskLimitTierModel
	ContractLiquidationModel    models.TContractLiquidationModel
	ContractInsuranceFundModel  models.TContractInsuranceFundAccountModel
	ContractFundingBatchModel   models.TContractFundingBatchModel
	ContractFundingSettleModel  models.TContractFundingSettlementModel
	FundingDifferenceAcctModel  models.TContractFundingDifferenceAccountModel
	ContractAdlExecutionModel   models.TContractAdlExecutionModel
	ContractDeliveryBatchModel  models.TContractDeliveryBatchModel
	ContractDeliverySettleModel models.TContractDeliverySettlementModel
	ContractUserConfigModel     models.TContractUserConfigModel
	TradeSymbolSecondsModel     models.TTradeSymbolSecondsModel
	TradeSymbolSessionModel     models.TTradeSymbolSessionModel
	ContractLeverageCfgModel    models.TContractLeverageConfigModel
	SymbolLeverageCfgModel      models.TTradeSymbolLeverageConfigModel
	SymbolLeverageDefaultModel  models.TTradeSymbolLeverageDefaultModel
	RiskUserTradeLimitModel     models.TRiskUserTradeLimitModel
	RiskUserSymbolLimitModel    models.TRiskUserSymbolLimitModel
	RiskOrderCheckLogModel      models.TRiskOrderCheckLogModel
	BizTradeEventModel          models.TBizTradeEventModel
	TradeEventInboxModel        models.TTradeEventInboxModel
	TradeAssetReservationModel  models.TTradeAssetReservationModel
	TradeSettlementInstrModel   models.TTradeSettlementInstructionModel
	AssetClient                 asset.AssetInternalClient
	MarketDataCache             *cache.MarketDataCache
	TradeMarketSnapshotModel    models.TTradeMarketSnapshotModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	marketRedis := v9.NewClient(&v9.Options{Addr: c.CacheRedis[0].Host, Username: c.CacheRedis[0].User, Password: c.CacheRedis[0].Pass})
	taskSubscriber := bus.NewSubscriberFromRedisConf(c.CacheRedis[0].RedisConf)
	tradeEventPublisher := bus.NewPublisherFromRedisConf(c.CacheRedis[0].RedisConf)
	tradeEventSubscriber := bus.NewSubscriberFromRedisConf(c.CacheRedis[0].RedisConf)
	hostname, _ := os.Hostname()
	instanceID := fmt.Sprintf("%s:%d:%d", hostname, os.Getpid(), time.Now().UnixNano())
	return &ServiceContext{
		Config:                      c,
		DB:                          conn,
		Redis:                       redis.MustNewRedis(c.Redis.RedisConf),
		TaskSubscriber:              taskSubscriber,
		TradeEventPublisher:         tradeEventPublisher,
		TradeEventSubscriber:        tradeEventSubscriber,
		TradeEventInstanceID:        instanceID,
		TradeSymbolModel:            models.NewTTradeSymbolModel(conn, c.CacheRedis),
		TradeSymbolSpotModel:        models.NewTTradeSymbolSpotModel(conn, c.CacheRedis),
		TradeSymbolContractModel:    models.NewTTradeSymbolContractModel(conn, c.CacheRedis),
		TradeUserConfigModel:        models.NewTTradeUserConfigModel(conn, c.CacheRedis),
		TradeOrderModel:             models.NewTTradeOrderModel(conn, c.CacheRedis),
		TradeOrderSpotModel:         models.NewTTradeOrderSpotModel(conn, c.CacheRedis),
		TradeOrderContractModel:     models.NewTTradeOrderContractModel(conn, c.CacheRedis),
		TradeOrderSecondsModel:      models.NewTTradeOrderSecondsModel(conn, c.CacheRedis),
		TradeSecondsPriceModel:      models.NewTTradeSecondsPriceSnapshotModel(conn, c.CacheRedis),
		TradeFillModel:              models.NewTTradeFillModel(conn, c.CacheRedis),
		TradeCancelLogModel:         models.NewTTradeCancelLogModel(conn, c.CacheRedis),
		ContractPositionModel:       models.NewTContractPositionModel(conn, c.CacheRedis),
		ContractPositionHistModel:   models.NewTContractPositionHistoryModel(conn, c.CacheRedis),
		ContractMarginSnapshotModel: models.NewTContractMarginSnapshotModel(conn, c.CacheRedis),
		ContractRiskLimitTierModel:  models.NewTContractRiskLimitTierModel(conn, c.CacheRedis),
		ContractLiquidationModel:    models.NewTContractLiquidationModel(conn, c.CacheRedis),
		ContractInsuranceFundModel:  models.NewTContractInsuranceFundAccountModel(conn, c.CacheRedis),
		ContractFundingBatchModel:   models.NewTContractFundingBatchModel(conn, c.CacheRedis),
		ContractFundingSettleModel:  models.NewTContractFundingSettlementModel(conn, c.CacheRedis),
		FundingDifferenceAcctModel:  models.NewTContractFundingDifferenceAccountModel(conn, c.CacheRedis),
		ContractAdlExecutionModel:   models.NewTContractAdlExecutionModel(conn, c.CacheRedis),
		ContractDeliveryBatchModel:  models.NewTContractDeliveryBatchModel(conn, c.CacheRedis),
		ContractDeliverySettleModel: models.NewTContractDeliverySettlementModel(conn, c.CacheRedis),
		ContractUserConfigModel:     models.NewTContractUserConfigModel(conn, c.CacheRedis),
		TradeSymbolSecondsModel:     models.NewTTradeSymbolSecondsModel(conn, c.CacheRedis),
		TradeSymbolSessionModel:     models.NewTTradeSymbolSessionModel(conn, c.CacheRedis),
		ContractLeverageCfgModel:    models.NewTContractLeverageConfigModel(conn, c.CacheRedis),
		SymbolLeverageCfgModel:      models.NewTTradeSymbolLeverageConfigModel(conn, c.CacheRedis),
		SymbolLeverageDefaultModel:  models.NewTTradeSymbolLeverageDefaultModel(conn, c.CacheRedis),
		RiskUserTradeLimitModel:     models.NewTRiskUserTradeLimitModel(conn, c.CacheRedis),
		RiskUserSymbolLimitModel:    models.NewTRiskUserSymbolLimitModel(conn, c.CacheRedis),
		RiskOrderCheckLogModel:      models.NewTRiskOrderCheckLogModel(conn, c.CacheRedis),
		BizTradeEventModel:          models.NewTBizTradeEventModel(conn, c.CacheRedis),
		TradeEventInboxModel:        models.NewTTradeEventInboxModel(conn, c.CacheRedis),
		TradeAssetReservationModel:  models.NewTTradeAssetReservationModel(conn, c.CacheRedis),
		TradeSettlementInstrModel:   models.NewTTradeSettlementInstructionModel(conn, c.CacheRedis),
		AssetClient:                 asset.NewAssetInternalClient(assetCli.Conn()),
		MarketDataCache:             cache.NewMarketDataCache(marketRedis),
		TradeMarketSnapshotModel:    models.NewTTradeMarketSnapshotModel(conn, c.CacheRedis),
	}
}

func (s *ServiceContext) GenerateBizNo(ctx context.Context, prefix string) (string, error) {
	now := time.Now()
	date := now.Format("20060102")

	// 每天、每个前缀单独计数
	key := fmt.Sprintf("order_id:%s:%s", prefix, date)

	seq, err := s.Redis.IncrCtx(ctx, key)
	if err != nil {
		return "", err
	}

	// 设置过期时间，避免 Redis 一直堆积旧 key
	// 这里只在第一次创建时设置
	if seq == 1 {
		_ = s.Redis.ExpireCtx(ctx, key, 36*int(time.Hour.Seconds()))
	}

	orderID := fmt.Sprintf("%s%s%06d", prefix, date, seq)
	return orderID, nil
}
