package svc

import (
	mq "wklive/common/mq/kafka"
	"wklive/proto/asset"
	"wklive/services/staking/internal/config"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config              config.Config
	DB                  sqlx.SqlConn
	Redis               *redis.Redis
	TaskSubscriber      *mq.Subscriber
	StakeOrderModel     models.TStakeOrderModel
	StakeProductModel   models.TStakeProductModel
	StakeRedeemLogModel models.TStakeRedeemLogModel
	StakeRewardLogModel models.TStakeRewardLogModel
	AssetClient         asset.InternalClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	mqConfig := mq.ForService(c.MQ, c.Name)
	taskSubscriber := mq.MustNewSubscriber(mqConfig, "staking-tasks")
	return &ServiceContext{
		Config:              c,
		DB:                  conn,
		Redis:               redis.MustNewRedis(c.Redis.RedisConf),
		TaskSubscriber:      taskSubscriber,
		StakeOrderModel:     models.NewTStakeOrderModel(conn, c.CacheRedis),
		StakeProductModel:   models.NewTStakeProductModel(conn, c.CacheRedis),
		StakeRedeemLogModel: models.NewTStakeRedeemLogModel(conn, c.CacheRedis),
		StakeRewardLogModel: models.NewTStakeRewardLogModel(conn, c.CacheRedis),
		AssetClient:         asset.NewInternalClient(assetCli.Conn()),
	}
}
