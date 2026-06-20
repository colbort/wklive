package svc

import (
	"context"
	"fmt"
	"time"
	"wklive/services/payment/internal/config"
	"wklive/services/payment/models"

	"wklive/proto/asset"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                     config.Config
	DB                         sqlx.SqlConn
	Redis                      *redis.Redis
	AssetCli                   asset.AssetInternalClient
	PayPlatformModel           models.TPayPlatformModel
	PayProductModel            models.TPayProductModel
	UserRechargeStatModel      models.TUserRechargeStatModel
	TenantPayAccountModel      models.TTenantPayAccountModel
	TenantPayChannelModel      models.TTenantPayChannelModel
	TenantPayChannelRuleModel  models.TTenantPayChannelRuleModel
	TenantPayPlatformModel     models.TTenantPayPlatformModel
	RechargeOrderModel         models.TRechargeOrderModel
	RechargeNotifyLogModel     models.TRechargeNotifyLogModel
	WithdrawOrderModel         models.TWithdrawOrderModel
	WithdrawNotifyLogModel     models.TWithdrawNotifyLogModel
	CryptoRechargeAddressModel models.TCryptoRechargeAddressModel
	CryptoWalletAccountModel   models.TCryptoWalletAccountModel
	CryptoRechargeTxModel      models.TCryptoRechargeTxModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	return &ServiceContext{
		Config:                     c,
		DB:                         conn,
		Redis:                      redis.MustNewRedis(c.Redis.RedisConf),
		AssetCli:                   asset.NewAssetInternalClient(assetCli.Conn()),
		PayPlatformModel:           models.NewTPayPlatformModel(conn, c.CacheRedis),
		PayProductModel:            models.NewTPayProductModel(conn, c.CacheRedis),
		UserRechargeStatModel:      models.NewTUserRechargeStatModel(conn, c.CacheRedis),
		TenantPayAccountModel:      models.NewTTenantPayAccountModel(conn, c.CacheRedis),
		TenantPayChannelModel:      models.NewTTenantPayChannelModel(conn, c.CacheRedis),
		TenantPayChannelRuleModel:  models.NewTTenantPayChannelRuleModel(conn, c.CacheRedis),
		TenantPayPlatformModel:     models.NewTTenantPayPlatformModel(conn, c.CacheRedis),
		RechargeOrderModel:         models.NewTRechargeOrderModel(conn, c.CacheRedis),
		RechargeNotifyLogModel:     models.NewTRechargeNotifyLogModel(conn, c.CacheRedis),
		WithdrawOrderModel:         models.NewTWithdrawOrderModel(conn, c.CacheRedis),
		WithdrawNotifyLogModel:     models.NewTWithdrawNotifyLogModel(conn, c.CacheRedis),
		CryptoRechargeAddressModel: models.NewTCryptoRechargeAddressModel(conn, c.CacheRedis),
		CryptoWalletAccountModel:   models.NewTCryptoWalletAccountModel(conn, c.CacheRedis),
		CryptoRechargeTxModel:      models.NewTCryptoRechargeTxModel(conn, c.CacheRedis),
	}
}

func (s *ServiceContext) GenerateOrderNo(ctx context.Context, prefix string) (string, error) {
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
