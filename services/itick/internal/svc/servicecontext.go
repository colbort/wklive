package svc

import (
	"time"
	"wklive/proto/system"
	"wklive/services/itick/internal/config"
	"wklive/services/itick/internal/klinewriter"
	"wklive/services/itick/internal/socket/client"
	"wklive/services/itick/internal/socket/server"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                   config.Config
	SystemCli                system.SystemClient
	ItickManager             *client.ItickManager
	Hub                      *server.Hub
	Factory                  *models.CoinKlineModelFactory
	Writer                   *klinewriter.BatchWriter
	ItickCategoryModel       models.ItickCategoryModel
	ItickProductModel        models.ItickProductModel
	ItickTenantCategoryModel models.ItickTenantCategoryModel
	ItickTenantProductModel  models.ItickTenantProductModel
	ItickSyncTaskModel       models.ItickSyncTaskModel
	ItickQuoteModel          models.ItickQuoteModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	hub := server.NewHub()

	conn := sqlx.NewMysql(c.Mysql.DataSource)
	itickCategoryModel := models.NewTItickCategoryModel(conn, c.CacheRedis).(models.ItickCategoryModel)
	itickManager := client.NewItickManager(c.Itick.WSUrl, c.Itick.Token, hub, itickCategoryModel)
	hub.SetHooks(
		func(key string, msg server.ClientMessage) {
			if err := itickManager.Subscribe(msg); err != nil {
				logx.Errorf("subscribe upstream failed, key=%s, err=%v", key, err)
			}
		},
		func(key string, msg server.ClientMessage) {
			if err := itickManager.Unsubscribe(msg); err != nil {
				logx.Errorf("unsubscribe upstream failed, key=%s, err=%v", key, err)
			}
		},
	)

	factory := models.NewCoinKlineModelFactory(c.Mongo.Url, c.Mongo.Db)

	writer := klinewriter.NewBatchWriter(
		factory,
		c.KlineWriter.QueueSize,
		c.KlineWriter.BatchSize,
		time.Duration(c.KlineWriter.FlushIntervalMs)*time.Millisecond,
		time.Duration(c.KlineWriter.WriteTimeoutMs)*time.Millisecond,
	)

	return &ServiceContext{
		Config:                   c,
		ItickManager:             itickManager,
		Hub:                      hub,
		Factory:                  factory,
		Writer:                   writer,
		ItickCategoryModel:       itickCategoryModel,
		ItickProductModel:        models.NewTItickProductModel(conn, c.CacheRedis).(models.ItickProductModel),
		ItickTenantCategoryModel: models.NewTItickTenantCategoryModel(conn, c.CacheRedis).(models.ItickTenantCategoryModel),
		ItickTenantProductModel:  models.NewTItickTenantProductModel(conn, c.CacheRedis).(models.ItickTenantProductModel),
		ItickSyncTaskModel:       models.NewTItickSyncTaskModel(conn, c.CacheRedis).(models.ItickSyncTaskModel),
		ItickQuoteModel:          models.NewTItickQuoteModel(conn, c.CacheRedis).(models.ItickQuoteModel),
	}
}
