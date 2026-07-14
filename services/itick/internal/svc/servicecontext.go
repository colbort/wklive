package svc

import (
	"context"
	"time"

	bus "wklive/common/bus/redis"
	"wklive/proto/option"
	"wklive/proto/system"
	"wklive/services/itick/internal/config"
	marketcalendar "wklive/services/itick/internal/market/calendar"
	"wklive/services/itick/internal/market/client"
	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/internal/pkg/klinewriter"
	"wklive/services/itick/models"

	icache "wklive/services/itick/internal/market/cache"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                      config.Config
	ItickRuntimeConfig          *system.ItickConfig
	SystemCli                   system.SystemClient
	OptionCli                   option.OptionInternalClient
	ItickManager                *client.ItickManager
	MarketDataCache             *icache.MarketDataCache
	BusRedis                    *redis.Client
	LockRedis                   *redis.Client
	TaskSubscriber              *bus.Subscriber
	Cache                       cache.Cache
	Factory                     *models.CoinKlineModelFactory
	Writer                      *klinewriter.BatchWriter
	RebuildDerivedKlines        func([]*models.CoinKline) error
	RebuildHistoricalKlines     func([]*models.CoinKline) error
	ItickCategoryModel          models.TItickCategoryModel
	ItickProductModel           models.TItickProductModel
	ItickTenantCategoryModel    models.TItickTenantCategoryModel
	ItickTenantProductModel     models.TItickTenantProductModel
	ItickSyncTaskModel          models.TItickSyncTaskModel
	ItickQuoteModel             models.TItickQuoteModel
	ItickKlineSyncProgressModel models.TItickKlineSyncProgressModel
	MarketCalendarModel         models.TItickMarketCalendarModel
	MarketCalendarResolver      *marketcalendar.Resolver
}

func NewServiceContext(c config.Config) *ServiceContext {
	systemCli := system.NewSystemClient(zrpc.MustNewClient(c.SystemRpc).Conn())
	optionCli := option.NewOptionInternalClient(zrpc.MustNewClient(c.OptionRpc).Conn())
	conn := sqlx.NewMysql(c.Mysql.DataSource)

	itickCategoryModel := models.NewTItickCategoryModel(conn, c.CacheRedis)
	itickProductModel := models.NewTItickProductModel(conn, c.CacheRedis)
	itickTenantCategoryModel := models.NewTItickTenantCategoryModel(conn, c.CacheRedis)
	itickTenantProductModel := models.NewTItickTenantProductModel(conn, c.CacheRedis)
	itickSyncTaskModel := models.NewTItickSyncTaskModel(conn, c.CacheRedis)
	itickQuoteModel := models.NewTItickQuoteModel(conn, c.CacheRedis)
	itickKlineSyncProgressModel := models.NewTItickKlineSyncProgressModel(conn, c.CacheRedis)
	marketCalendarModel := models.NewTItickMarketCalendarModel(conn, c.CacheRedis)
	marketCalendarResolver := marketcalendar.NewResolver(marketCalendarModel, 10*time.Minute)

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
	marketDataCache := icache.NewMarketDataCache(busRedis)
	taskSubscriber := bus.NewSubscriberFromRedisConf(c.CacheRedis[0].RedisConf)

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
		c.Itick.ApiUrl,
		c.Itick.Token,
		itickCategoryModel,
		itickProductModel,
		busRedis,
		lockRedis,
		marketDataCache,
	)
	itickManager.SetQuoteHandler(func(_ context.Context, msg types.ClientMessage, payload *types.QuotePayload) {
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

	return &ServiceContext{
		Config:                      c,
		SystemCli:                   systemCli,
		OptionCli:                   optionCli,
		ItickManager:                itickManager,
		MarketDataCache:             marketDataCache,
		BusRedis:                    busRedis,
		LockRedis:                   lockRedis,
		TaskSubscriber:              taskSubscriber,
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
		MarketCalendarModel:         marketCalendarModel,
		MarketCalendarResolver:      marketCalendarResolver,
	}
}
