package svc

import (
	"wklive/proto/trade"
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
	ProviderAdapters       *provider.Registry
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	tradeClient := zrpc.MustNewClient(c.TradeRpc)
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
		ProviderAdapters:       provider.NewRegistry(),
	}
}
