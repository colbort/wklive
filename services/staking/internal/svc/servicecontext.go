package svc

import (
	"wklive/proto/asset"
	"wklive/services/staking/internal/config"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config              config.Config
	DB                  sqlx.SqlConn
	StakeOrderModel     models.StakeOrderModel
	StakeProductModel   models.StakeProductModel
	StakeRedeemLogModel models.StakeRedeemLogModel
	StakeRewardLogModel models.StakeRewardLogModel
	AssetClient         asset.AssetInternalClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	return &ServiceContext{
		Config:              c,
		DB:                  conn,
		StakeOrderModel:     models.NewTStakeOrderModel(conn, c.CacheRedis).(models.StakeOrderModel),
		StakeProductModel:   models.NewTStakeProductModel(conn, c.CacheRedis).(models.StakeProductModel),
		StakeRedeemLogModel: models.NewTStakeRedeemLogModel(conn, c.CacheRedis).(models.StakeRedeemLogModel),
		StakeRewardLogModel: models.NewTStakeRewardLogModel(conn, c.CacheRedis).(models.StakeRewardLogModel),
		AssetClient:         asset.NewAssetInternalClient(assetCli.Conn()),
	}
}
