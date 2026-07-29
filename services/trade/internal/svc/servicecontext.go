package svc

import (
	"fmt"
	"os"
	"time"
	cache "wklive/common/market"
	mq "wklive/common/mq/kafka"
	"wklive/services/trade/internal/config"
	"wklive/services/trade/internal/delayqueue"
	"wklive/services/trade/models"

	"wklive/proto/asset"
	"wklive/proto/itick"

	v9 "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                       config.Config
	TransactionModel             models.TransactionModel
	Redis                        *redis.Redis
	TaskSubscriber               *mq.Subscriber
	TradeEventPublisher          *mq.Publisher
	TradeEventSubscriber         *mq.Subscriber
	TradeEventInstanceID         string
	TradeSymbolModel             models.TTradeSymbolModel
	TradeSymbolSpotModel         models.TTradeSymbolSpotModel
	TradeSymbolContractModel     models.TTradeSymbolContractModel
	TradeUserConfigModel         models.TTradeUserConfigModel
	TradeOrderModel              models.TTradeOrderModel
	TradeOrderSpotModel          models.TTradeOrderSpotModel
	TradeOrderContractModel      models.TTradeOrderContractModel
	TradeOrderSecondsModel       models.TTradeOrderSecondsModel
	TradeSecondsPriceModel       models.TTradeSecondsPriceSnapshotModel
	TradeFillModel               models.TTradeFillModel
	TradeCancelLogModel          models.TTradeCancelLogModel
	ContractPositionModel        models.TContractPositionModel
	ContractPositionHistModel    models.TContractPositionHistoryModel
	ContractMarginSnapshotModel  models.TContractMarginSnapshotModel
	ContractAccountLiqModel      models.TContractAccountLiquidationModel
	ContractAccountLiqItemModel  models.TContractAccountLiquidationItemModel
	ContractRiskLimitTierModel   models.TContractRiskLimitTierModel
	ContractLiquidationModel     models.TContractLiquidationModel
	ContractInsuranceFundModel   models.TContractInsuranceFundAccountModel
	ContractFundingBatchModel    models.TContractFundingBatchModel
	ContractFundingSettleModel   models.TContractFundingSettlementModel
	FundingDifferenceAcctModel   models.TContractFundingDifferenceAccountModel
	ContractAdlExecutionModel    models.TContractAdlExecutionModel
	ContractDeliveryBatchModel   models.TContractDeliveryBatchModel
	ContractDeliverySettleModel  models.TContractDeliverySettlementModel
	ContractUserConfigModel      models.TContractUserConfigModel
	TradeSymbolSecondsModel      models.TTradeSymbolSecondsModel
	TradeSymbolSessionModel      models.TTradeSymbolSessionModel
	ContractLeverageCfgModel     models.TContractLeverageConfigModel
	SymbolLeverageCfgModel       models.TTradeSymbolLeverageConfigModel
	SymbolLeverageDefaultModel   models.TTradeSymbolLeverageDefaultModel
	RiskUserTradeLimitModel      models.TRiskUserTradeLimitModel
	RiskUserSymbolLimitModel     models.TRiskUserSymbolLimitModel
	RiskOrderCheckLogModel       models.TRiskOrderCheckLogModel
	BizTradeEventModel           models.TBizTradeEventModel
	TradeEventInboxModel         models.TTradeEventInboxModel
	TradeAssetReservationModel   models.TTradeAssetReservationModel
	TradeSettlementInstrModel    models.TTradeSettlementInstructionModel
	ContractReconcileIssueModel  models.TContractReconciliationIssueModel
	ContractReconcileCursorModel models.TContractReconciliationCursorModel
	AssetClient                  asset.AssetClient
	AssetAdminClient             asset.AdminClient
	ItickClient                  itick.ItickClient
	MarketDataCache              *cache.MarketDataCache
	TradeMarketSnapshotModel     models.TTradeMarketSnapshotModel
	DelayQueue                   *delayqueue.Queue
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	itickCli := zrpc.MustNewClient(c.ItickRpc)
	marketRedis := v9.NewClient(&v9.Options{Addr: c.CacheRedis[0].Host, Username: c.CacheRedis[0].User, Password: c.CacheRedis[0].Pass})
	mqConfig := mq.ForService(c.MQ, c.Name)
	taskSubscriber := mq.MustNewSubscriber(mqConfig, "trade-tasks")
	tradeEventPublisher := mq.MustNewPublisher(mqConfig)
	tradeEventSubscriber := mq.MustNewSubscriber(mqConfig, "trade-realtime")
	hostname, _ := os.Hostname()
	instanceID := fmt.Sprintf("%s:%d:%d", hostname, os.Getpid(), time.Now().UnixNano())
	delayQueue, err := delayqueue.New(c.DelayQueue.Enabled, c.DelayQueue.Beanstalks, c.Redis.RedisConf)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:                       c,
		TransactionModel:             models.NewTransactionModel(conn, c.CacheRedis),
		Redis:                        redis.MustNewRedis(c.Redis.RedisConf),
		TaskSubscriber:               taskSubscriber,
		TradeEventPublisher:          tradeEventPublisher,
		TradeEventSubscriber:         tradeEventSubscriber,
		TradeEventInstanceID:         instanceID,
		TradeSymbolModel:             models.NewTTradeSymbolModel(conn, c.CacheRedis),
		TradeSymbolSpotModel:         models.NewTTradeSymbolSpotModel(conn, c.CacheRedis),
		TradeSymbolContractModel:     models.NewTTradeSymbolContractModel(conn, c.CacheRedis),
		TradeUserConfigModel:         models.NewTTradeUserConfigModel(conn, c.CacheRedis),
		TradeOrderModel:              models.NewTTradeOrderModel(conn, c.CacheRedis),
		TradeOrderSpotModel:          models.NewTTradeOrderSpotModel(conn, c.CacheRedis),
		TradeOrderContractModel:      models.NewTTradeOrderContractModel(conn, c.CacheRedis),
		TradeOrderSecondsModel:       models.NewTTradeOrderSecondsModel(conn, c.CacheRedis),
		TradeSecondsPriceModel:       models.NewTTradeSecondsPriceSnapshotModel(conn, c.CacheRedis),
		TradeFillModel:               models.NewTTradeFillModel(conn, c.CacheRedis),
		TradeCancelLogModel:          models.NewTTradeCancelLogModel(conn, c.CacheRedis),
		ContractPositionModel:        models.NewTContractPositionModel(conn, c.CacheRedis),
		ContractPositionHistModel:    models.NewTContractPositionHistoryModel(conn, c.CacheRedis),
		ContractMarginSnapshotModel:  models.NewTContractMarginSnapshotModel(conn, c.CacheRedis),
		ContractAccountLiqModel:      models.NewTContractAccountLiquidationModel(conn, c.CacheRedis),
		ContractAccountLiqItemModel:  models.NewTContractAccountLiquidationItemModel(conn, c.CacheRedis),
		ContractRiskLimitTierModel:   models.NewTContractRiskLimitTierModel(conn, c.CacheRedis),
		ContractLiquidationModel:     models.NewTContractLiquidationModel(conn, c.CacheRedis),
		ContractInsuranceFundModel:   models.NewTContractInsuranceFundAccountModel(conn, c.CacheRedis),
		ContractFundingBatchModel:    models.NewTContractFundingBatchModel(conn, c.CacheRedis),
		ContractFundingSettleModel:   models.NewTContractFundingSettlementModel(conn, c.CacheRedis),
		FundingDifferenceAcctModel:   models.NewTContractFundingDifferenceAccountModel(conn, c.CacheRedis),
		ContractAdlExecutionModel:    models.NewTContractAdlExecutionModel(conn, c.CacheRedis),
		ContractDeliveryBatchModel:   models.NewTContractDeliveryBatchModel(conn, c.CacheRedis),
		ContractDeliverySettleModel:  models.NewTContractDeliverySettlementModel(conn, c.CacheRedis),
		ContractUserConfigModel:      models.NewTContractUserConfigModel(conn, c.CacheRedis),
		TradeSymbolSecondsModel:      models.NewTTradeSymbolSecondsModel(conn, c.CacheRedis),
		TradeSymbolSessionModel:      models.NewTTradeSymbolSessionModel(conn, c.CacheRedis),
		ContractLeverageCfgModel:     models.NewTContractLeverageConfigModel(conn, c.CacheRedis),
		SymbolLeverageCfgModel:       models.NewTTradeSymbolLeverageConfigModel(conn, c.CacheRedis),
		SymbolLeverageDefaultModel:   models.NewTTradeSymbolLeverageDefaultModel(conn, c.CacheRedis),
		RiskUserTradeLimitModel:      models.NewTRiskUserTradeLimitModel(conn, c.CacheRedis),
		RiskUserSymbolLimitModel:     models.NewTRiskUserSymbolLimitModel(conn, c.CacheRedis),
		RiskOrderCheckLogModel:       models.NewTRiskOrderCheckLogModel(conn, c.CacheRedis),
		BizTradeEventModel:           models.NewTBizTradeEventModel(conn, c.CacheRedis),
		TradeEventInboxModel:         models.NewTTradeEventInboxModel(conn, c.CacheRedis),
		TradeAssetReservationModel:   models.NewTTradeAssetReservationModel(conn, c.CacheRedis),
		TradeSettlementInstrModel:    models.NewTTradeSettlementInstructionModel(conn, c.CacheRedis),
		ContractReconcileIssueModel:  models.NewTContractReconciliationIssueModel(conn, c.CacheRedis),
		ContractReconcileCursorModel: models.NewTContractReconciliationCursorModel(conn, c.CacheRedis),
		AssetClient:                  asset.NewAssetClient(assetCli.Conn()),
		AssetAdminClient:             asset.NewAdminClient(assetCli.Conn()),
		ItickClient:                  itick.NewItickClient(itickCli.Conn()),
		MarketDataCache:              cache.NewMarketDataCache(marketRedis),
		TradeMarketSnapshotModel:     models.NewTTradeMarketSnapshotModel(conn, c.CacheRedis),
		DelayQueue:                   delayQueue,
	}
}
