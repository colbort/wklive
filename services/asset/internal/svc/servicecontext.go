package svc

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wklive/services/asset/internal/config"
	"wklive/services/asset/models"

	cache "wklive/common/market"

	v9 "github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config               config.Config
	DB                   sqlx.SqlConn
	Redis                *redis.Redis
	UserAssetModel       models.TUserAssetModel
	AssetLockModel       models.TAssetLockModel
	AssetFlowModel       models.TAssetFlowModel
	AssetFreezeModel     models.TAssetFreezeModel
	AssetIdempotentModel models.TAssetIdempotentModel
	AssetCoinConfigModel models.TAssetCoinConfigModel
	MarketDataCache      *cache.MarketDataCache
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	marketRedis := v9.NewClient(&v9.Options{Addr: c.CacheRedis[0].Host, Username: c.CacheRedis[0].User, Password: c.CacheRedis[0].Pass})
	return &ServiceContext{
		Config:               c,
		DB:                   conn,
		Redis:                redis.MustNewRedis(c.Redis.RedisConf),
		UserAssetModel:       models.NewTUserAssetModel(conn, c.CacheRedis),
		AssetLockModel:       models.NewTAssetLockModel(conn, c.CacheRedis),
		AssetFlowModel:       models.NewTAssetFlowModel(conn, c.CacheRedis),
		AssetFreezeModel:     models.NewTAssetFreezeModel(conn, c.CacheRedis),
		AssetIdempotentModel: models.NewTAssetIdempotentModel(conn, c.CacheRedis),
		AssetCoinConfigModel: models.NewTAssetCoinConfigModel(conn, c.CacheRedis),
		MarketDataCache:      cache.NewMarketDataCache(marketRedis),
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

// 获取最新报价
func (s *ServiceContext) LastPrice(ctx context.Context, symbol string) (decimal.Decimal, error) {
	msg := cache.NormalizeClientMessage(cache.ClientMessage{
		Topic:        cache.TopicQuote,
		CategoryCode: "crypto",
		Market:       "BA",
		Symbol:       symbol,
	})
	items, err := s.MarketDataCache.ReadMany(ctx, []cache.ClientMessage{msg})
	if err != nil {
		return decimal.NewFromInt(0), err
	}
	if len(items) == 0 {
		return decimal.NewFromInt(0), redis.Nil
	}
	data, ok := items[0].Payload.(*cache.QuotePayload)
	if !ok || data == nil {
		return decimal.NewFromInt(0), redis.Nil
	}
	return decimal.NewFromFloat(data.LastPrice), nil
}
