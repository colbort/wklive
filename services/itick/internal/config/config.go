package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	CacheRedis cache.CacheConf
	BusRedis   cache.CacheConf
	LockRedis  cache.CacheConf
	Mysql      struct {
		DataSource string
	} `json:"Mysql" yaml:"Mysql"`
	Itick ItickConf
	Mongo struct {
		Url string
		Db  string
	}
	KlineWriter struct {
		// QueueSize 是 K 线写入缓冲队列容量。
		// 当队列满时，新进入的数据会被 Enqueue 拒绝。
		QueueSize int

		// BatchSize 是单次批量写入的最大条数。
		// 按 categoryCode + interval 分桶后，单个桶达到该数量会立即 flush。
		BatchSize int

		// FlushIntervalMs 是定时刷盘间隔，单位毫秒。
		// 即使未达到 BatchSize，也会在这个周期触发一次批量写入。
		FlushIntervalMs int

		// WriteTimeoutMs 是单次 MongoDB 批量写入超时时间，单位毫秒。
		WriteTimeoutMs int
	}
}

type ItickConf struct {
	ApiUrl string
	WSUrl  string
	Token  string
}
