package config

import (
	mq "wklive/common/mq/kafka"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	CacheRedis cache.CacheConf
	MQ         mq.Config
	DataCache  redis.RedisKeyConf
	LockRedis  redis.RedisKeyConf
	Mysql      struct {
		DataSource string
	} `json:"Mysql" yaml:"Mysql"`
	Market  MarketConf
	Runtime MarketRuntimeConf
	// ExternalQuotes are independent public market-data producers used by the
	// contract price engine. They publish only FINAL_QUOTE snapshots.
	ExternalQuotes []ExternalQuoteSourceConf
	Mongo          struct {
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
	// SnapshotOutbox 控制权威快照异步发布吞吐。
	SnapshotOutbox struct {
		// WorkerCount 是单实例并发发布协程数。
		WorkerCount int
		// BatchSize 是单轮最多领取的任务数。
		BatchSize int64
		// IdleIntervalMs 是队列暂时为空时的轮询间隔。
		IdleIntervalMs int
	}
	// SnapshotOutboxCleanup 控制已成功发布的权威快照 Outbox 清理任务。
	// 未配置或配置为非正数时，任务使用代码内置的安全默认值。
	SnapshotOutboxCleanup struct {
		// SuccessRetentionMinutes 是 status=3 成功记录的保留时间，单位分钟。
		SuccessRetentionMinutes int

		// IntervalSeconds 是清理任务的执行周期，单位秒。
		IntervalSeconds int

		// BatchSize 是单条 DELETE 语句最多删除的记录数，上限为 10000。
		BatchSize int64

		// MaxBatchesPerRun 是每轮任务最多执行的删除批次数，用于限制单轮数据库压力。
		MaxBatchesPerRun int

		// BatchPauseMs 是相邻删除批次之间的等待时间，单位毫秒。
		BatchPauseMs int
	}
	// AuthoritativeCache controls the bounded Redis authoritative snapshot cache.
	AuthoritativeCache struct {
		// HotWindowMinutes is the Redis historical lookback window. Older reads use MySQL.
		HotWindowMinutes int
		// LegacyCleanupEnabled removes the obsolete v1/v2 Redis layout after a coordinated rollout.
		LegacyCleanupEnabled bool
		// LegacyCleanupScanCount limits each non-blocking SCAN batch.
		LegacyCleanupScanCount int64
		// LegacyCleanupIntervalSeconds is the delay between legacy cleanup batches.
		LegacyCleanupIntervalSeconds int
	}
}

type ExternalQuoteSourceConf struct {
	Enabled         bool
	Authority       string
	ProviderCode    string
	Adapter         string
	BaseURL         string
	CategoryCode    string
	Market          string
	Symbol          string
	UpstreamSymbol  string
	IntervalMs      int
	TimeoutMs       int
	MaxSourceAgeMs  int64
	MaxFutureSkewMs int64
}

type MarketRuntimeConf struct {
	ReconcileIntervalMinutes int64
	ReconcileWindowBars      int64
	GapScanIntervalMinutes   int64
	RepairBatchSize          int64
	BuildingBucketTtlMinutes int64
	WsKlineStaleSeconds      int64
}

type MarketConf struct {
	ApiUrl                 string
	WSUrl                  string
	Token                  string
	RestRateLimitPerMinute int
	RestRateLimitBurst     int
}
