package config

import (
	"wklive/common/mq/kafka"

	"github.com/zeromicro/go-queue/dq"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	CacheRedis cache.CacheConf
	MQ         mq.Config
	Mysql      struct {
		DataSource string
	} `json:"Mysql" yaml:"Mysql"`
	AssetRpc   zrpc.RpcClientConf
	DelayQueue struct {
		Enabled    bool
		Beanstalks []dq.Beanstalk
	}
	MarketSnapshotInboxCleanup struct {
		RetentionHours   int
		IntervalMinutes  int
		BatchSize        int64
		MaxBatchesPerRun int
	}
	InsuranceInventoryExit struct {
		Enabled               bool
		MaxQuantityPerRequest string
		MaxPremiumPerRequest  string
		MaxDailyQuantity      string
		MaxMarkDeviationRatio string
		MinOrderBookQuantity  string
	}
	PlatformBackstop struct {
		// Enabled is an emergency admission and execution gate only. It does
		// not replace the monetary limits that must be enforced by Asset.
		Enabled bool
	}
	ProductScope ProductScope
}

// ProductScope contains the production admission switches for optional product
// capabilities. The zero value is deliberately fail-closed.
type ProductScope struct {
	SellerTradingEnabled           bool
	PortfolioMarginEnabled         bool
	PhysicalDeliveryEnabled        bool
	ComplexOrdersEnabled           bool
	PublicMarketEnabled            bool
	MMPEnabled                     bool
	AmericanExerciseEnabled        bool
}
