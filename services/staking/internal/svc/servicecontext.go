package svc

import (
	"context"
	"fmt"
	"time"
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
		Redis:               redis.MustNewRedis(c.Redis.RedisConf),
		StakeOrderModel:     models.NewTStakeOrderModel(conn, c.CacheRedis).(models.StakeOrderModel),
		StakeProductModel:   models.NewTStakeProductModel(conn, c.CacheRedis).(models.StakeProductModel),
		StakeRedeemLogModel: models.NewTStakeRedeemLogModel(conn, c.CacheRedis).(models.StakeRedeemLogModel),
		StakeRewardLogModel: models.NewTStakeRewardLogModel(conn, c.CacheRedis).(models.StakeRewardLogModel),
		AssetClient:         asset.NewAssetInternalClient(assetCli.Conn()),
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
