package svc

import (
	"context"
	"time"

	"wklive/proto/option"
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
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                      config.Config
	SystemCli                   system.SystemClient
	OptionCli                   option.OptionInternalClient
	ItickManager                *client.ItickManager
	Hub                         *server.Hub
	LockRedis                   *redis.Client
	Cache                       cache.Cache
	Factory                     *models.CoinKlineModelFactory
	Writer                      *klinewriter.BatchWriter
	ItickCategoryModel          models.TItickCategoryModel
	ItickProductModel           models.TItickProductModel
	ItickTenantCategoryModel    models.TItickTenantCategoryModel
	ItickTenantProductModel     models.TItickTenantProductModel
	ItickSyncTaskModel          models.TItickSyncTaskModel
	ItickQuoteModel             models.TItickQuoteModel
	ItickKlineSyncProgressModel models.TItickKlineSyncProgressModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	systemCli := system.NewSystemClient(zrpc.MustNewClient(c.SystemRpc).Conn())
	optionCli := option.NewOptionInternalClient(zrpc.MustNewClient(c.OptionRpc).Conn())
	hub := server.NewHub()

	conn := sqlx.NewMysql(c.Mysql.DataSource)

	itickCategoryModel := models.NewTItickCategoryModel(conn, c.CacheRedis)
	itickProductModel := models.NewTItickProductModel(conn, c.CacheRedis)
	itickTenantCategoryModel := models.NewTItickTenantCategoryModel(conn, c.CacheRedis)
	itickTenantProductModel := models.NewTItickTenantProductModel(conn, c.CacheRedis)
	itickSyncTaskModel := models.NewTItickSyncTaskModel(conn, c.CacheRedis)
	itickQuoteModel := models.NewTItickQuoteModel(conn, c.CacheRedis)
	itickKlineSyncProgressModel := models.NewTItickKlineSyncProgressModel(conn, c.CacheRedis)

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
	itickManager.SetQuoteHandler(func(_ context.Context, msg server.ClientMessage, payload *client.QuotePayload) {
		rpcCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		resp, err := optionCli.SyncMarketQuote(rpcCtx, &option.SyncMarketQuoteReq{
			CategoryCode:    msg.CategoryCode,
			Market:          msg.Market,
			Symbol:          msg.Symbol,
			UnderlyingPrice: payload.LastPrice,
			OpenPrice:       payload.Open,
			HighPrice:       payload.High,
			LowPrice:        payload.Low,
			Volume:          payload.Volume,
			Turnover:        payload.Turnover,
			QuoteTs:         payload.Ts,
		})
		if err != nil {
			logx.Errorf("sync option market quote failed, symbol=%s market=%s err=%v", msg.Symbol, msg.Market, err)
			return
		}
		if resp == nil || resp.GetBase() == nil {
			logx.Errorf("sync option market quote empty response, symbol=%s market=%s", msg.Symbol, msg.Market)
			return
		}
		if resp.GetBase().GetCode() != 200 {
			logx.Errorf("sync option market quote rejected, symbol=%s market=%s code=%d msg=%s",
				msg.Symbol, msg.Market, resp.GetBase().GetCode(), resp.GetBase().GetMsg())
		}
	})

	ctx := context.Background()

	if err := itickManager.Load(ctx); err != nil {
		logx.Errorf("itick manager load failed: %v", err)
	}

	if err := itickManager.Start(ctx); err != nil {
		logx.Errorf("itick manager start failed: %v", err)
	}
	return &ServiceContext{
		Config:                      c,
		SystemCli:                   systemCli,
		OptionCli:                   optionCli,
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
