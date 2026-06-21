package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Jwt struct {
		AccessSecret string
		AccessExpire int64
	} `json:"Jwt" yaml:"Jwt"`
	CacheRedis cache.CacheConf
	Mysql      struct {
		DataSource string
	} `json:"Mysql" yaml:"Mysql"`
	Mongo struct {
		Url string
		Db  string
	}
}
