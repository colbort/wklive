package config

import (
	"errors"
	"strings"
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

	TradeRpc  zrpc.RpcClientConf
	AssetRpc  zrpc.RpcClientConf
	MarketRpc zrpc.RpcClientConf
	UserRpc   zrpc.RpcClientConf

	MarketAuthorities    []string
	PriceEngineAuthority string

	OKX struct {
		Enabled   bool
		BaseURL   string
		TimeoutMs int64
	}
}

func (c Config) Validate() error {
	for _, authority := range c.MarketAuthorities {
		if strings.TrimSpace(authority) != "" {
			return nil
		}
	}
	return errors.New("MarketAuthorities must contain at least one authority")
}
