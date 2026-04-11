package svc

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wklive/services/asset/internal/config"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config               config.Config
	Redis                *redis.Redis
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
		Redis:                redis.MustNewRedis(c.Redis.RedisConf),
		UserAssetModel:       models.NewTUserAssetModel(conn, c.CacheRedis).(models.UserAssetModel),
		AssetLockModel:       models.NewTAssetLockModel(conn, c.CacheRedis).(models.AssetLockModel),
		AssetFlowModel:       models.NewTAssetFlowModel(conn, c.CacheRedis).(models.AssetFlowModel),
		AssetFreezeModel:     models.NewTAssetFreezeModel(conn, c.CacheRedis).(models.AssetFreezeModel),
		AssetIdempotentModel: models.NewTAssetIdempotentModel(conn, c.CacheRedis).(models.AssetIdempotentModel),
	}
}

func (s *ServiceContext) GenerateOrderNo(ctx context.Context, prefix string, bizNo string) (string, error) {
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
	if bizNo != "" {
		return fmt.Sprintf("%s_%s", orderID, SanitizeBizNo(bizNo)), nil
	}
	return orderID, nil
}

func SanitizeBizNo(bizNo string) string {
	return strings.Map(func(r rune) rune {
		if r == '_' || r == '-' || r == '.' || r == '/' || r == ':' {
			return '_'
		}
		if r >= '0' && r <= '9' {
			return r
		}
		if r >= 'a' && r <= 'z' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r
		}
		return -1
	}, bizNo)
}
