package svc

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wklive/common/i18n"
	"wklive/proto/itick"
	"wklive/services/asset/internal/config"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
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
	ItickClient          itick.ItickAppClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	itickCli := zrpc.MustNewClient(c.ItickRpc)
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
		ItickClient:          itick.NewItickAppClient(itickCli.Conn()),
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
	resp, err := s.ItickClient.GetQuote(ctx, &itick.GetQuoteReq{
		CategoryCode: "crypto",
		Market:       "BA",
		Symbol:       symbol,
	})
	if err != nil {
		return decimal.Zero, err
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		return decimal.Zero, i18n.StatusError(ctx, i18n.InvalidExchangeRate)
	}
	price, err := decimal.NewFromString(resp.GetData().GetLastPrice())
	if err != nil || !price.IsPositive() {
		return decimal.Zero, i18n.StatusError(ctx, i18n.InvalidExchangeRate)
	}
	return price, nil
}
