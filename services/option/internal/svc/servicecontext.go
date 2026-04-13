package svc

import (
	"context"
	"fmt"
	"time"
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
		Redis:                     redis.MustNewRedis(c.Redis.RedisConf),
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

func (s *ServiceContext) GenerateBizNo(ctx context.Context, prefix string) (string, error) {
	now := time.Now()
	date := now.Format("20060102")

	// 每天、每个前缀单独计数
	key := fmt.Sprintf("order_id:%s:%s", prefix, date)

	seq, err := s.Redis.IncrCtx(ctx, key)
	if err != nil {
		return "", err
	}

	// 设置过期时间，避免 Redis 一直堆积旧 key
	// 这里只在第一次创建时设置
	if seq == 1 {
		_ = s.Redis.ExpireCtx(ctx, key, 36*int(time.Hour.Seconds()))
	}

	orderID := fmt.Sprintf("%s%s%06d", prefix, date, seq)
	return orderID, nil
}
