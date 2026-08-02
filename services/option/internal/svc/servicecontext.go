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
	Config                              config.Config
	DB                                  sqlx.SqlConn
	Redis                               *redis.Redis
	TaskSubscriber                      *mq.Subscriber
	UserEventPublisher                  *mq.Publisher
	MarketSnapshotSubscriber            *mq.Subscriber
	OptionContractModel                 models.TOptionContractModel
	OptionMarketModel                   models.TOptionMarketModel
	OptionMarketSnapshotModel           models.TOptionMarketSnapshotModel
	OptionMarketSnapshotInboxModel      models.TOptionMarketSnapshotInboxModel
	OptionOrderModel                    models.TOptionOrderModel
	OptionComboOrderModel               models.TOptionComboOrderModel
	OptionComboOrderLegModel            models.TOptionComboOrderLegModel
	OptionTradeModel                    models.TOptionTradeModel
	OptionPositionModel                 models.TOptionPositionModel
	OptionExerciseModel                 models.TOptionExerciseModel
	OptionExerciseAssignmentModel       models.TOptionExerciseAssignmentModel
	OptionExerciseInstructionModel      models.TOptionExerciseInstructionModel
	OptionMmpConfigModel                models.TOptionMmpConfigModel
	OptionUserTradingControlModel       models.TOptionUserTradingControlModel
	OptionTradingControlEventModel      models.TOptionTradingControlEventModel
	OptionTradeCorrectionModel          models.TOptionTradeCorrectionModel
	OptionTradeCorrectionLegModel       models.TOptionTradeCorrectionLegModel
	OptionSettlementModel               models.TOptionSettlementModel
	OptionSettlementPriceModel          models.TOptionSettlementPriceModel
	OptionSettlementBatchModel          models.TOptionSettlementBatchModel
	OptionSettlementDetailModel         models.TOptionSettlementDetailModel
	OptionPhysicalDeliveryUnitModel     models.TOptionPhysicalDeliveryUnitModel
	OptionTradingCalendarModel          models.TOptionTradingCalendarModel
	OptionTradingCalendarSessionModel   models.TOptionTradingCalendarSessionModel
	OptionTradingCalendarExceptionModel models.TOptionTradingCalendarExceptionModel
	OptionTradingHaltModel              models.TOptionTradingHaltModel
	OptionCorporateActionModel          models.TOptionCorporateActionModel
	OptionCorporateActionContractModel  models.TOptionCorporateActionContractModel
	OptionCorporateActionPositionModel  models.TOptionCorporateActionPositionModel
	OptionCorporateActionMarginLotModel models.TOptionCorporateActionMarginLotModel
	OptionContractSeriesModel           models.TOptionContractSeriesModel
	OptionContractSeriesExpiryModel     models.TOptionContractSeriesExpiryModel
	OptionContractSeriesStrikeBandModel models.TOptionContractSeriesStrikeBandModel
	OptionContractSeriesDetailModel     models.TOptionContractSeriesDetailModel
	OptionReconciliationIssueModel      models.TOptionReconciliationIssueModel
	OptionAccountModel                  models.TOptionAccountModel
	OptionBillModel                     models.TOptionBillModel
	OptionAssetInstructionModel         models.TOptionAssetInstructionModel
	OptionOutboxModel                   models.TOptionOutboxModel
	OptionInboxModel                    models.TOptionInboxModel
	OptionMarginLotModel                models.TOptionMarginLotModel
	OptionRiskAccountModel              models.TOptionRiskAccountModel
	OptionPortfolioRiskConfigModel      models.TOptionPortfolioRiskConfigModel
	OptionLiquidationModel              models.TOptionLiquidationModel
	OptionInsuranceInventoryExitModel   models.TOptionInsuranceInventoryExitModel
	OptionInsuranceFundFlowModel        models.TOptionInsuranceFundFlowModel
	AssetClient                         asset.AssetClient
	DelayQueue                          *delayqueue.Queue
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
		Config:                              c,
		DB:                                  conn,
		Redis:                               redis.MustNewRedis(c.Redis.RedisConf),
		TaskSubscriber:                      taskSubscriber,
		UserEventPublisher:                  userEventPublisher,
		MarketSnapshotSubscriber:            marketSnapshotSubscriber,
		OptionContractModel:                 models.NewTOptionContractModel(conn, c.CacheRedis),
		OptionMarketModel:                   models.NewTOptionMarketModel(conn, c.CacheRedis),
		OptionMarketSnapshotModel:           models.NewTOptionMarketSnapshotModel(conn, c.CacheRedis),
		OptionMarketSnapshotInboxModel:      models.NewTOptionMarketSnapshotInboxModel(conn, c.CacheRedis),
		OptionOrderModel:                    models.NewTOptionOrderModel(conn, c.CacheRedis),
		OptionComboOrderModel:               models.NewTOptionComboOrderModel(conn, c.CacheRedis),
		OptionComboOrderLegModel:            models.NewTOptionComboOrderLegModel(conn, c.CacheRedis),
		OptionTradeModel:                    models.NewTOptionTradeModel(conn, c.CacheRedis),
		OptionPositionModel:                 models.NewTOptionPositionModel(conn, c.CacheRedis),
		OptionExerciseModel:                 models.NewTOptionExerciseModel(conn, c.CacheRedis),
		OptionExerciseAssignmentModel:       models.NewTOptionExerciseAssignmentModel(conn, c.CacheRedis),
		OptionExerciseInstructionModel:      models.NewTOptionExerciseInstructionModel(conn, c.CacheRedis),
		OptionMmpConfigModel:                models.NewTOptionMmpConfigModel(conn, c.CacheRedis),
		OptionUserTradingControlModel:       models.NewTOptionUserTradingControlModel(conn, c.CacheRedis),
		OptionTradingControlEventModel:      models.NewTOptionTradingControlEventModel(conn, c.CacheRedis),
		OptionTradeCorrectionModel:          models.NewTOptionTradeCorrectionModel(conn, c.CacheRedis),
		OptionTradeCorrectionLegModel:       models.NewTOptionTradeCorrectionLegModel(conn, c.CacheRedis),
		OptionSettlementModel:               models.NewTOptionSettlementModel(conn, c.CacheRedis),
		OptionSettlementPriceModel:          models.NewTOptionSettlementPriceModel(conn, c.CacheRedis),
		OptionSettlementBatchModel:          models.NewTOptionSettlementBatchModel(conn, c.CacheRedis),
		OptionSettlementDetailModel:         models.NewTOptionSettlementDetailModel(conn, c.CacheRedis),
		OptionPhysicalDeliveryUnitModel:     models.NewTOptionPhysicalDeliveryUnitModel(conn, c.CacheRedis),
		OptionTradingCalendarModel:          models.NewTOptionTradingCalendarModel(conn, c.CacheRedis),
		OptionTradingCalendarSessionModel:   models.NewTOptionTradingCalendarSessionModel(conn, c.CacheRedis),
		OptionTradingCalendarExceptionModel: models.NewTOptionTradingCalendarExceptionModel(conn, c.CacheRedis),
		OptionTradingHaltModel:              models.NewTOptionTradingHaltModel(conn, c.CacheRedis),
		OptionCorporateActionModel:          models.NewTOptionCorporateActionModel(conn, c.CacheRedis),
		OptionCorporateActionContractModel:  models.NewTOptionCorporateActionContractModel(conn, c.CacheRedis),
		OptionCorporateActionPositionModel:  models.NewTOptionCorporateActionPositionModel(conn, c.CacheRedis),
		OptionCorporateActionMarginLotModel: models.NewTOptionCorporateActionMarginLotModel(conn, c.CacheRedis),
		OptionContractSeriesModel:           models.NewTOptionContractSeriesModel(conn, c.CacheRedis),
		OptionContractSeriesExpiryModel:     models.NewTOptionContractSeriesExpiryModel(conn, c.CacheRedis),
		OptionContractSeriesStrikeBandModel: models.NewTOptionContractSeriesStrikeBandModel(conn, c.CacheRedis),
		OptionContractSeriesDetailModel:     models.NewTOptionContractSeriesDetailModel(conn, c.CacheRedis),
		OptionReconciliationIssueModel:      models.NewTOptionReconciliationIssueModel(conn, c.CacheRedis),
		OptionAccountModel:                  models.NewTOptionAccountModel(conn, c.CacheRedis),
		OptionBillModel:                     models.NewTOptionBillModel(conn, c.CacheRedis),
		OptionAssetInstructionModel:         models.NewTOptionAssetInstructionModel(conn, c.CacheRedis),
		OptionOutboxModel:                   models.NewTOptionOutboxModel(conn, c.CacheRedis),
		OptionInboxModel:                    models.NewTOptionInboxModel(conn, c.CacheRedis),
		OptionMarginLotModel:                models.NewTOptionMarginLotModel(conn, c.CacheRedis),
		OptionRiskAccountModel:              models.NewTOptionRiskAccountModel(conn, c.CacheRedis),
		OptionPortfolioRiskConfigModel:      models.NewTOptionPortfolioRiskConfigModel(conn, c.CacheRedis),
		OptionLiquidationModel:              models.NewTOptionLiquidationModel(conn, c.CacheRedis),
		OptionInsuranceInventoryExitModel:   models.NewTOptionInsuranceInventoryExitModel(conn, c.CacheRedis),
		OptionInsuranceFundFlowModel:        models.NewTOptionInsuranceFundFlowModel(conn, c.CacheRedis),
		AssetClient:                         asset.NewAssetClient(assetCli.Conn()),
		DelayQueue:                          queue,
	}
}
