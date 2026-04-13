package svc

import (
	"wklive/proto/asset"
	"wklive/services/option/internal/config"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                    config.Config
	DB                        sqlx.SqlConn
	OptionContractModel       models.OptionContractModel
	OptionMarketModel         models.OptionMarketModel
	OptionMarketSnapshotModel models.OptionMarketSnapshotModel
	OptionOrderModel          models.OptionOrderModel
	OptionTradeModel          models.OptionTradeModel
	OptionPositionModel       models.OptionPositionModel
	OptionExerciseModel       models.OptionExerciseModel
	OptionSettlementModel     models.OptionSettlementModel
	OptionAccountModel        models.OptionAccountModel
	OptionBillModel           models.OptionBillModel
	AssetClient               asset.AssetInternalClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	return &ServiceContext{
		Config:                    c,
		DB:                        conn,
		OptionContractModel:       models.NewTOptionContractModel(conn, c.CacheRedis).(models.OptionContractModel),
		OptionMarketModel:         models.NewTOptionMarketModel(conn, c.CacheRedis).(models.OptionMarketModel),
		OptionMarketSnapshotModel: models.NewTOptionMarketSnapshotModel(conn, c.CacheRedis).(models.OptionMarketSnapshotModel),
		OptionOrderModel:          models.NewTOptionOrderModel(conn, c.CacheRedis).(models.OptionOrderModel),
		OptionTradeModel:          models.NewTOptionTradeModel(conn, c.CacheRedis).(models.OptionTradeModel),
		OptionPositionModel:       models.NewTOptionPositionModel(conn, c.CacheRedis).(models.OptionPositionModel),
		OptionExerciseModel:       models.NewTOptionExerciseModel(conn, c.CacheRedis).(models.OptionExerciseModel),
		OptionSettlementModel:     models.NewTOptionSettlementModel(conn, c.CacheRedis).(models.OptionSettlementModel),
		OptionAccountModel:        models.NewTOptionAccountModel(conn, c.CacheRedis).(models.OptionAccountModel),
		OptionBillModel:           models.NewTOptionBillModel(conn, c.CacheRedis).(models.OptionBillModel),
		AssetClient:               asset.NewAssetInternalClient(assetCli.Conn()),
	}
}
