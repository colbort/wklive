package svc

import (
	"context"
	"fmt"
	"time"
	"wklive/services/trade/internal/config"
	"wklive/services/trade/models"

	"wklive/proto/asset"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                    config.Config
	DB                        sqlx.SqlConn
	Redis                     *redis.Redis
	TradeSymbolModel          models.TradeSymbolModel
	TradeSymbolSpotModel      models.TradeSymbolSpotModel
	TradeSymbolContractModel  models.TradeSymbolContractModel
	TradeUserConfigModel      models.TradeUserConfigModel
	TradeOrderModel           models.TradeOrderModel
	TradeOrderSpotModel       models.TradeOrderSpotModel
	TradeOrderContractModel   models.TradeOrderContractModel
	TradeFillModel            models.TradeFillModel
	TradeCancelLogModel       models.TradeCancelLogModel
	ContractPositionModel     models.ContractPositionModel
	ContractPositionHistModel models.ContractPositionHistoryModel
	ContractMarginAcctModel   models.ContractMarginAccountModel
	ContractLeverageCfgModel  models.ContractLeverageConfigModel
	RiskUserTradeLimitModel   models.RiskUserTradeLimitModel
	RiskUserSymbolLimitModel  models.RiskUserSymbolLimitModel
	RiskOrderCheckLogModel    models.RiskOrderCheckLogModel
	BizTradeEventModel        models.BizTradeEventModel
	AssetClient               asset.AssetInternalClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	return &ServiceContext{
		Config:                    c,
		DB:                        conn,
		Redis:                     redis.MustNewRedis(c.Redis.RedisConf),
		TradeSymbolModel:          models.NewTTradeSymbolModel(conn, c.CacheRedis).(models.TradeSymbolModel),
		TradeSymbolSpotModel:      models.NewTTradeSymbolSpotModel(conn, c.CacheRedis).(models.TradeSymbolSpotModel),
		TradeSymbolContractModel:  models.NewTTradeSymbolContractModel(conn, c.CacheRedis).(models.TradeSymbolContractModel),
		TradeUserConfigModel:      models.NewTTradeUserConfigModel(conn, c.CacheRedis).(models.TradeUserConfigModel),
		TradeOrderModel:           models.NewTTradeOrderModel(conn, c.CacheRedis).(models.TradeOrderModel),
		TradeOrderSpotModel:       models.NewTTradeOrderSpotModel(conn, c.CacheRedis).(models.TradeOrderSpotModel),
		TradeOrderContractModel:   models.NewTTradeOrderContractModel(conn, c.CacheRedis).(models.TradeOrderContractModel),
		TradeFillModel:            models.NewTTradeFillModel(conn, c.CacheRedis).(models.TradeFillModel),
		TradeCancelLogModel:       models.NewTTradeCancelLogModel(conn, c.CacheRedis).(models.TradeCancelLogModel),
		ContractPositionModel:     models.NewTContractPositionModel(conn, c.CacheRedis).(models.ContractPositionModel),
		ContractPositionHistModel: models.NewTContractPositionHistoryModel(conn, c.CacheRedis).(models.ContractPositionHistoryModel),
		ContractMarginAcctModel:   models.NewTContractMarginAccountModel(conn, c.CacheRedis).(models.ContractMarginAccountModel),
		ContractLeverageCfgModel:  models.NewTContractLeverageConfigModel(conn, c.CacheRedis).(models.ContractLeverageConfigModel),
		RiskUserTradeLimitModel:   models.NewTRiskUserTradeLimitModel(conn, c.CacheRedis).(models.RiskUserTradeLimitModel),
		RiskUserSymbolLimitModel:  models.NewTRiskUserSymbolLimitModel(conn, c.CacheRedis).(models.RiskUserSymbolLimitModel),
		RiskOrderCheckLogModel:    models.NewTRiskOrderCheckLogModel(conn, c.CacheRedis).(models.RiskOrderCheckLogModel),
		BizTradeEventModel:        models.NewTBizTradeEventModel(conn, c.CacheRedis).(models.BizTradeEventModel),
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
