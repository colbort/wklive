package svc

import (
	mq "wklive/common/mq/kafka"
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
	MQPublisher                *mq.Publisher
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
		MQPublisher:                mq.MustNewPublisher(mq.ForService(c.MQ, c.Name)),
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
