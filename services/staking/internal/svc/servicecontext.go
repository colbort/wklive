package svc

import (
	mq "wklive/common/mq/kafka"
	"wklive/proto/asset"
	"wklive/services/staking/internal/config"
	"wklive/services/staking/internal/delayqueue"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                   config.Config
	DB                       sqlx.SqlConn
	Redis                    *redis.Redis
	TaskSubscriber           *mq.Subscriber
	UserEventPublisher       *mq.Publisher
	StakeOrderModel          models.TStakeOrderModel
	StakeProductModel        models.TStakeProductModel
	StakeRedeemLogModel      models.TStakeRedeemLogModel
	StakeRewardLogModel      models.TStakeRewardLogModel
	StakeOperationModel      models.TStakeOperationModel
	StakeReconciliationModel models.TStakeReconciliationModel
	StakeUserPositionModel   models.TStakeUserPositionModel
	AssetClient              asset.AssetClient
	AssetAdminClient         asset.AdminClient
	DelayQueue               *delayqueue.Queue
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	mqConfig := mq.ForService(c.MQ, c.Name)
	taskSubscriber := mq.MustNewSubscriber(mqConfig, "staking-tasks")
	userEventPublisher := mq.MustNewPublisher(mqConfig)
	queue, err := delayqueue.New(c.DelayQueue.Enabled, c.DelayQueue.Beanstalks, c.Redis.RedisConf)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:                   c,
		DB:                       conn,
		Redis:                    redis.MustNewRedis(c.Redis.RedisConf),
		TaskSubscriber:           taskSubscriber,
		UserEventPublisher:       userEventPublisher,
		StakeOrderModel:          models.NewTStakeOrderModel(conn, c.CacheRedis),
		StakeProductModel:        models.NewTStakeProductModel(conn, c.CacheRedis),
		StakeRedeemLogModel:      models.NewTStakeRedeemLogModel(conn, c.CacheRedis),
		StakeRewardLogModel:      models.NewTStakeRewardLogModel(conn, c.CacheRedis),
		StakeOperationModel:      models.NewTStakeOperationModel(conn, c.CacheRedis),
		StakeReconciliationModel: models.NewTStakeReconciliationModel(conn, c.CacheRedis),
		StakeUserPositionModel:   models.NewTStakeUserPositionModel(conn, c.CacheRedis),
		AssetClient:              asset.NewAssetClient(assetCli.Conn()),
		AssetAdminClient:         asset.NewAdminClient(assetCli.Conn()),
		DelayQueue:               queue,
	}
}
