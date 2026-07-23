package svc

import (
	mq "wklive/common/mq/kafka"
	"wklive/proto/asset"
	"wklive/services/option/internal/config"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                    config.Config
	DB                        sqlx.SqlConn
	Redis                     *redis.Redis
	TaskSubscriber            *mq.Subscriber
	OptionContractModel       models.TOptionContractModel
	OptionMarketModel         models.TOptionMarketModel
	OptionMarketSnapshotModel models.TOptionMarketSnapshotModel
	OptionOrderModel          models.TOptionOrderModel
	OptionTradeModel          models.TOptionTradeModel
	OptionPositionModel       models.TOptionPositionModel
	OptionExerciseModel       models.TOptionExerciseModel
	OptionSettlementModel     models.TOptionSettlementModel
	OptionAccountModel        models.TOptionAccountModel
	OptionBillModel           models.TOptionBillModel
	AssetClient               asset.InternalClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	mqConfig := mq.ForService(c.MQ, c.Name)
	taskSubscriber := mq.MustNewSubscriber(mqConfig, "option-tasks")
	return &ServiceContext{
		Config:                    c,
		DB:                        conn,
		Redis:                     redis.MustNewRedis(c.Redis.RedisConf),
		TaskSubscriber:            taskSubscriber,
		OptionContractModel:       models.NewTOptionContractModel(conn, c.CacheRedis),
		OptionMarketModel:         models.NewTOptionMarketModel(conn, c.CacheRedis),
		OptionMarketSnapshotModel: models.NewTOptionMarketSnapshotModel(conn, c.CacheRedis),
		OptionOrderModel:          models.NewTOptionOrderModel(conn, c.CacheRedis),
		OptionTradeModel:          models.NewTOptionTradeModel(conn, c.CacheRedis),
		OptionPositionModel:       models.NewTOptionPositionModel(conn, c.CacheRedis),
		OptionExerciseModel:       models.NewTOptionExerciseModel(conn, c.CacheRedis),
		OptionSettlementModel:     models.NewTOptionSettlementModel(conn, c.CacheRedis),
		OptionAccountModel:        models.NewTOptionAccountModel(conn, c.CacheRedis),
		OptionBillModel:           models.NewTOptionBillModel(conn, c.CacheRedis),
		AssetClient:               asset.NewInternalClient(assetCli.Conn()),
	}
}
