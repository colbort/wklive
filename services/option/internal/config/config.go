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
}
