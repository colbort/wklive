package svc

import (
	market "wklive/common/market"
	mq "wklive/common/mq/kafka"
	"wklive/proto/asset"
	"wklive/services/option/internal/config"
	"wklive/services/option/internal/delayqueue"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                         config.Config
	DB                             sqlx.SqlConn
	Redis                          *redis.Redis
	TaskSubscriber                 *mq.Subscriber
	UserEventPublisher             *mq.Publisher
	MarketSnapshotSubscriber       *mq.Subscriber
	OptionContractModel            models.TOptionContractModel
	OptionMarketModel              models.TOptionMarketModel
	OptionMarketSnapshotModel      models.TOptionMarketSnapshotModel
	OptionMarketSnapshotInboxModel models.TOptionMarketSnapshotInboxModel
	OptionOrderModel               models.TOptionOrderModel
	OptionTradeModel               models.TOptionTradeModel
	OptionPositionModel            models.TOptionPositionModel
	OptionExerciseModel            models.TOptionExerciseModel
	OptionExerciseAssignmentModel  models.TOptionExerciseAssignmentModel
	OptionSettlementModel          models.TOptionSettlementModel
	OptionSettlementPriceModel     models.TOptionSettlementPriceModel
	OptionSettlementBatchModel     models.TOptionSettlementBatchModel
	OptionSettlementDetailModel    models.TOptionSettlementDetailModel
	OptionReconciliationIssueModel models.TOptionReconciliationIssueModel
	OptionAccountModel             models.TOptionAccountModel
	OptionBillModel                models.TOptionBillModel
	OptionAssetInstructionModel    models.TOptionAssetInstructionModel
	OptionOutboxModel              models.TOptionOutboxModel
	OptionInboxModel               models.TOptionInboxModel
	OptionMarginLotModel           models.TOptionMarginLotModel
	OptionRiskAccountModel         models.TOptionRiskAccountModel
	OptionLiquidationModel         models.TOptionLiquidationModel
	OptionInsuranceFundFlowModel   models.TOptionInsuranceFundFlowModel
	AssetClient                    asset.AssetClient
	DelayQueue                     *delayqueue.Queue
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	mqConfig := mq.ForService(c.MQ, c.Name)
	taskSubscriber := mq.MustNewSubscriber(mqConfig, "option-tasks")
	userEventPublisher := mq.MustNewPublisher(mqConfig)
	marketSnapshotSubscriber := mq.MustNewSubscriber(mqConfig, market.OptionMarketQuoteConsumerGroup)
	queue, err := delayqueue.New(c.DelayQueue.Enabled, c.DelayQueue.Beanstalks, c.Redis.RedisConf)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:                         c,
		DB:                             conn,
		Redis:                          redis.MustNewRedis(c.Redis.RedisConf),
		TaskSubscriber:                 taskSubscriber,
		UserEventPublisher:             userEventPublisher,
		MarketSnapshotSubscriber:       marketSnapshotSubscriber,
		OptionContractModel:            models.NewTOptionContractModel(conn, c.CacheRedis),
		OptionMarketModel:              models.NewTOptionMarketModel(conn, c.CacheRedis),
		OptionMarketSnapshotModel:      models.NewTOptionMarketSnapshotModel(conn, c.CacheRedis),
		OptionMarketSnapshotInboxModel: models.NewTOptionMarketSnapshotInboxModel(conn, c.CacheRedis),
		OptionOrderModel:               models.NewTOptionOrderModel(conn, c.CacheRedis),
		OptionTradeModel:               models.NewTOptionTradeModel(conn, c.CacheRedis),
		OptionPositionModel:            models.NewTOptionPositionModel(conn, c.CacheRedis),
		OptionExerciseModel:            models.NewTOptionExerciseModel(conn, c.CacheRedis),
		OptionExerciseAssignmentModel:  models.NewTOptionExerciseAssignmentModel(conn, c.CacheRedis),
		OptionSettlementModel:          models.NewTOptionSettlementModel(conn, c.CacheRedis),
		OptionSettlementPriceModel:     models.NewTOptionSettlementPriceModel(conn, c.CacheRedis),
		OptionSettlementBatchModel:     models.NewTOptionSettlementBatchModel(conn, c.CacheRedis),
		OptionSettlementDetailModel:    models.NewTOptionSettlementDetailModel(conn, c.CacheRedis),
		OptionReconciliationIssueModel: models.NewTOptionReconciliationIssueModel(conn, c.CacheRedis),
		OptionAccountModel:             models.NewTOptionAccountModel(conn, c.CacheRedis),
		OptionBillModel:                models.NewTOptionBillModel(conn, c.CacheRedis),
		OptionAssetInstructionModel:    models.NewTOptionAssetInstructionModel(conn, c.CacheRedis),
		OptionOutboxModel:              models.NewTOptionOutboxModel(conn, c.CacheRedis),
		OptionInboxModel:               models.NewTOptionInboxModel(conn, c.CacheRedis),
		OptionMarginLotModel:           models.NewTOptionMarginLotModel(conn, c.CacheRedis),
		OptionRiskAccountModel:         models.NewTOptionRiskAccountModel(conn, c.CacheRedis),
		OptionLiquidationModel:         models.NewTOptionLiquidationModel(conn, c.CacheRedis),
		OptionInsuranceFundFlowModel:   models.NewTOptionInsuranceFundFlowModel(conn, c.CacheRedis),
		AssetClient:                    asset.NewAssetClient(assetCli.Conn()),
		DelayQueue:                     queue,
	}
}
