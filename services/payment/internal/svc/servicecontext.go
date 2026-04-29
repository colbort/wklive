package svc

import (
	"context"
	"fmt"
	"time"
	"wklive/services/payment/internal/config"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                     config.Config
	DB                         sqlx.SqlConn
	Redis                      *redis.Redis
	PayPlatformModel           models.PayPlatformModel
	PayProductModel            models.PayProductModel
	UserRechargeStatModel      models.UserRechargeStatModel
	TenantPayAccountModel      models.TenantPayAccountModel
	TenantPayChannelModel      models.TenantPayChannelModel
	TenantPayChannelRuleModel  models.TenantPayChannelRuleModel
	TenantPayPlatformModel     models.TenantPayPlatformModel
	RechargeOrderModel         models.RechargeOrderModel
	RechargeNotifyLogModel     models.RechargeNotifyLogModel
	WithdrawOrderModel         models.WithdrawOrderModel
	WithdrawNotifyLogModel     models.WithdrawNotifyLogModel
	CryptoRechargeAddressModel models.CryptoRechargeAddressModel
	CryptoWalletAccountModel   models.CryptoWalletAccountModel
	CryptoRechargeTxModel      models.CryptoRechargeTxModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:                     c,
		DB:                         conn,
		Redis:                      redis.MustNewRedis(c.Redis.RedisConf),
		PayPlatformModel:           models.NewTPayPlatformModel(conn, c.CacheRedis).(models.PayPlatformModel),
		PayProductModel:            models.NewTPayProductModel(conn, c.CacheRedis).(models.PayProductModel),
		UserRechargeStatModel:      models.NewTUserRechargeStatModel(conn, c.CacheRedis).(models.UserRechargeStatModel),
		TenantPayAccountModel:      models.NewTTenantPayAccountModel(conn, c.CacheRedis).(models.TenantPayAccountModel),
		TenantPayChannelModel:      models.NewTTenantPayChannelModel(conn, c.CacheRedis).(models.TenantPayChannelModel),
		TenantPayChannelRuleModel:  models.NewTTenantPayChannelRuleModel(conn, c.CacheRedis).(models.TenantPayChannelRuleModel),
		TenantPayPlatformModel:     models.NewTTenantPayPlatformModel(conn, c.CacheRedis).(models.TenantPayPlatformModel),
		RechargeOrderModel:         models.NewTRechargeOrderModel(conn, c.CacheRedis).(models.RechargeOrderModel),
		RechargeNotifyLogModel:     models.NewTRechargeNotifyLogModel(conn, c.CacheRedis).(models.RechargeNotifyLogModel),
		WithdrawOrderModel:         models.NewTWithdrawOrderModel(conn, c.CacheRedis).(models.WithdrawOrderModel),
		WithdrawNotifyLogModel:     models.NewTWithdrawNotifyLogModel(conn, c.CacheRedis).(models.WithdrawNotifyLogModel),
		CryptoRechargeAddressModel: models.NewTCryptoRechargeAddressModel(conn, c.CacheRedis).(models.CryptoRechargeAddressModel),
		CryptoWalletAccountModel:   models.NewTCryptoWalletAccountModel(conn, c.CacheRedis).(models.CryptoWalletAccountModel),
		CryptoRechargeTxModel:      models.NewTCryptoRechargeTxModel(conn, c.CacheRedis).(models.CryptoRechargeTxModel),
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
