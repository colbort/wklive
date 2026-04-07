package svc

import (
	"context"
	"time"

	"wklive/proto/system"
	"wklive/services/itick/internal/config"
	"wklive/services/itick/internal/pkg/klinewriter"
	"wklive/services/itick/internal/socket/client"
	"wklive/services/itick/internal/socket/server"
	"wklive/services/itick/models"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
)

type ServiceContext struct {
	Config                      config.Config
	SystemCli                   system.SystemClient
	ItickManager                *client.ItickManager
	Hub                         *server.Hub
	LockRedis                   *redis.Client
	Cache                       cache.Cache
	Factory                     *models.CoinKlineModelFactory
	Writer                      *klinewriter.BatchWriter
	ItickCategoryModel          models.ItickCategoryModel
	ItickProductModel           models.ItickProductModel
	ItickTenantCategoryModel    models.ItickTenantCategoryModel
	ItickTenantProductModel     models.ItickTenantProductModel
	ItickSyncTaskModel          models.ItickSyncTaskModel
	ItickQuoteModel             models.ItickQuoteModel
	ItickKlineSyncProgressModel models.ItickKlineSyncProgressModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	hub := server.NewHub()

	conn := sqlx.NewMysql(c.Mysql.DataSource)

	itickCategoryModel := models.NewTItickCategoryModel(conn, c.CacheRedis).(models.ItickCategoryModel)
	itickProductModel := models.NewTItickProductModel(conn, c.CacheRedis).(models.ItickProductModel)
	itickTenantCategoryModel := models.NewTItickTenantCategoryModel(conn, c.CacheRedis).(models.ItickTenantCategoryModel)
	itickTenantProductModel := models.NewTItickTenantProductModel(conn, c.CacheRedis).(models.ItickTenantProductModel)
	itickSyncTaskModel := models.NewTItickSyncTaskModel(conn, c.CacheRedis).(models.ItickSyncTaskModel)
	itickQuoteModel := models.NewTItickQuoteModel(conn, c.CacheRedis).(models.ItickQuoteModel)
	itickKlineSyncProgressModel := models.NewTItickKlineSyncProgressModel(conn, c.CacheRedis).(models.ItickKlineSyncProgressModel)

	busRedis := redis.NewClient(&redis.Options{
		Addr:     c.BusRedis[0].Host,
		Username: c.BusRedis[0].User,
		Password: c.BusRedis[0].Pass,
		DB:       0,
	})

	lockRedis := redis.NewClient(&redis.Options{
		Addr:     c.LockRedis[0].Host,
		Username: c.LockRedis[0].User,
		Password: c.LockRedis[0].Pass,
		DB:       0,
	})

	// 这里不能 defer Close，不然函数返回后 Redis 连接就被关掉了
	// defer rdb.Close()

	factory := models.NewCoinKlineModelFactory(c.Mongo.Url, c.Mongo.Db)

	writer := klinewriter.NewBatchWriter(
		factory,
		c.KlineWriter.QueueSize,
		c.KlineWriter.BatchSize,
		time.Duration(c.KlineWriter.FlushIntervalMs)*time.Millisecond,
		time.Duration(c.KlineWriter.WriteTimeoutMs)*time.Millisecond,
	)

	itickManager := client.NewItickManager(
		c.Itick.WSUrl,
		c.Itick.Token,
		hub,
		itickCategoryModel,
		busRedis,
		lockRedis,
	)

	ctx := context.Background()

	if err := itickManager.Load(ctx); err != nil {
		logx.Errorf("itick manager load failed: %v", err)
	}

	if err := itickManager.Start(ctx); err != nil {
		logx.Errorf("itick manager start failed: %v", err)
	}
	cache.New(c.CacheRedis, syncx.NewSingleFlight(), cache.NewStat(""), redis.Nil)
	return &ServiceContext{
		Config:                      c,
		ItickManager:                itickManager,
		Hub:                         hub,
		LockRedis:                   lockRedis,
		Cache:                       cache.New(c.CacheRedis, syncx.NewSingleFlight(), cache.NewStat("quote"), redis.Nil),
		Factory:                     factory,
		Writer:                      writer,
		ItickCategoryModel:          itickCategoryModel,
		ItickProductModel:           itickProductModel,
		ItickTenantCategoryModel:    itickTenantCategoryModel,
		ItickTenantProductModel:     itickTenantProductModel,
		ItickSyncTaskModel:          itickSyncTaskModel,
		ItickQuoteModel:             itickQuoteModel,
		ItickKlineSyncProgressModel: itickKlineSyncProgressModel,
	}
}
