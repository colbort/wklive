package svc

import (
	"wklive/services/asset/internal/config"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config               config.Config
	UserAssetModel       models.UserAssetModel
	AssetLockModel       models.AssetLockModel
	AssetFlowModel       models.AssetFlowModel
	AssetFreezeModel     models.AssetFreezeModel
	AssetIdempotentModel models.AssetIdempotentModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:               c,
		UserAssetModel:       models.NewTUserAssetModel(conn, c.CacheRedis).(models.UserAssetModel),
		AssetLockModel:       models.NewTAssetLockModel(conn, c.CacheRedis).(models.AssetLockModel),
		AssetFlowModel:       models.NewTAssetFlowModel(conn, c.CacheRedis).(models.AssetFlowModel),
		AssetFreezeModel:     models.NewTAssetFreezeModel(conn, c.CacheRedis).(models.AssetFreezeModel),
		AssetIdempotentModel: models.NewTAssetIdempotentModel(conn, c.CacheRedis).(models.AssetIdempotentModel),
	}
}
