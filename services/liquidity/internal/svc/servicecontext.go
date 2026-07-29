package svc

import (
	"time"

	mq "wklive/common/mq/kafka"
	"wklive/proto/asset"
	"wklive/proto/market"
	"wklive/proto/trade"
	"wklive/proto/user"
	"wklive/services/liquidity/internal/config"
	"wklive/services/liquidity/internal/provider"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	ProviderModel          models.TLiquidityProviderModel
	SymbolConfigModel      models.TLiquiditySymbolConfigModel
	StrategyLevelModel     models.TLiquidityStrategyLevelModel
	QuoteCycleModel        models.TLiquidityQuoteCycleModel
	QuoteOrderModel        models.TLiquidityQuoteOrderModel
	ExternalOrderModel     models.TLiquidityExternalOrderModel
	ExternalFillModel      models.TLiquidityExternalFillModel
	HedgeTaskModel         models.TLiquidityHedgeTaskModel
	InventorySnapshotModel models.TLiquidityInventorySnapshotModel
	RiskEventModel         models.TLiquidityRiskEventModel
	ReconcileBatchModel    models.TLiquidityReconcileBatchModel
	ReconcileDetailModel   models.TLiquidityReconcileDetailModel
	EventInboxModel        models.TLiquidityEventInboxModel
	EventOutboxModel       models.TLiquidityEventOutboxModel
	TradeClient            trade.TradeClient
	MarketClient            market.MarketClient
	UserClient             user.UserClient
	AssetClient            asset.AssetClient
	InternalMarketMaker    provider.InternalMarketMaker
	ProviderAdapters       *provider.Registry
	TaskSubscriber         *mq.Subscriber
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	tradeClient := zrpc.MustNewClient(c.TradeRpc)
	marketClient := zrpc.MustNewClient(c.MarketRpc)
	userClient := zrpc.MustNewClient(c.UserRpc)
	assetClient := zrpc.MustNewClient(c.AssetRpc)
	mqConfig := mq.ForService(c.MQ, c.Name)
	providerAdapters := provider.NewRegistry()
	if err := providerAdapters.Register("OKX", provider.NewOKXAdapter(
		c.OKX.Enabled,
		provider.EnvCredentialResolver{},
		c.OKX.BaseURL,
		time.Duration(c.OKX.TimeoutMs)*time.Millisecond,
	)); err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:                 c,
		ProviderModel:          models.NewTLiquidityProviderModel(conn, c.CacheRedis),
		SymbolConfigModel:      models.NewTLiquiditySymbolConfigModel(conn, c.CacheRedis),
		StrategyLevelModel:     models.NewTLiquidityStrategyLevelModel(conn, c.CacheRedis),
		QuoteCycleModel:        models.NewTLiquidityQuoteCycleModel(conn, c.CacheRedis),
		QuoteOrderModel:        models.NewTLiquidityQuoteOrderModel(conn, c.CacheRedis),
		ExternalOrderModel:     models.NewTLiquidityExternalOrderModel(conn, c.CacheRedis),
		ExternalFillModel:      models.NewTLiquidityExternalFillModel(conn, c.CacheRedis),
		HedgeTaskModel:         models.NewTLiquidityHedgeTaskModel(conn, c.CacheRedis),
		InventorySnapshotModel: models.NewTLiquidityInventorySnapshotModel(conn, c.CacheRedis),
		RiskEventModel:         models.NewTLiquidityRiskEventModel(conn, c.CacheRedis),
		ReconcileBatchModel:    models.NewTLiquidityReconcileBatchModel(conn, c.CacheRedis),
		ReconcileDetailModel:   models.NewTLiquidityReconcileDetailModel(conn, c.CacheRedis),
		EventInboxModel:        models.NewTLiquidityEventInboxModel(conn, c.CacheRedis),
		EventOutboxModel:       models.NewTLiquidityEventOutboxModel(conn, c.CacheRedis),
		TradeClient:            trade.NewTradeClient(tradeClient.Conn()),
		MarketClient:            market.NewMarketClient(marketClient.Conn()),
		UserClient:             user.NewUserClient(userClient.Conn()),
		AssetClient:            asset.NewAssetClient(assetClient.Conn()),
		InternalMarketMaker:    provider.NewTradeInternalMarketMaker(trade.NewTradeClient(tradeClient.Conn())),
		ProviderAdapters:       providerAdapters,
		TaskSubscriber:         mq.MustNewSubscriber(mqConfig, "liquidity-tasks"),
	}
}
