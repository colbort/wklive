package config

import (
	mq "wklive/common/mq/kafka"

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
	AssetRpc             zrpc.RpcClientConf
	MarketRpc             zrpc.RpcClientConf
	MarketAuthority      string
	PriceEngineAuthority string
	DelayQueue           struct {
		Enabled    bool
		Beanstalks []dq.Beanstalk
	}
	LiquidityOrderArchive struct {
		Enabled       bool
		RetentionDays int64
		BatchSize     int64
	}
	AutomaticLiquidation struct {
		// Enabled must remain false until the P1-02 liquidation, insurance-fund
		// and ADL production gate has been fully accepted.
		Enabled bool
	}
	CrossMarginTrading struct {
		// Enabled allows only exposure-increasing CROSS orders. Reduce-only
		// exits remain available even while this production gate is closed.
		Enabled bool
	}
}
