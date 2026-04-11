package svc

import (
	"wklive/services/staking/internal/config"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config              config.Config
	StakeOrderModel     models.StakeOrderModel
	StakeProductModel   models.StakeProductModel
	StakeRedeemLogModel models.StakeRedeemLogModel
	StakeRewardLogModel models.StakeRewardLogModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:              c,
		StakeOrderModel:     models.NewTStakeOrderModel(conn, c.CacheRedis).(models.StakeOrderModel),
		StakeProductModel:   models.NewTStakeProductModel(conn, c.CacheRedis).(models.StakeProductModel),
		StakeRedeemLogModel: models.NewTStakeRedeemLogModel(conn, c.CacheRedis).(models.StakeRedeemLogModel),
		StakeRewardLogModel: models.NewTStakeRewardLogModel(conn, c.CacheRedis).(models.StakeRewardLogModel),
	}
}
