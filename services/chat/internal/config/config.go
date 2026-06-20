package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	CacheRedis cache.CacheConf
	BusRedis   redis.RedisConf `json:"BusRedis" yaml:"BusRedis"`
	Mysql      struct {
		DataSource string
	} `json:"Mysql" yaml:"Mysql"`
	Mongo struct {
		Url string
		Db  string
	}
}
