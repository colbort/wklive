package svc

import (
	"context"
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
