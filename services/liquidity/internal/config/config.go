package config

import (
	mq "wklive/common/mq/kafka"

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

	TradeRpc zrpc.RpcClientConf
	AssetRpc zrpc.RpcClientConf
	ItickRpc zrpc.RpcClientConf
	UserRpc  zrpc.RpcClientConf

	MarketAuthority      string
	PriceEngineAuthority string

	OKX struct {
		Enabled   bool
		BaseURL   string
		TimeoutMs int64
	}
}
