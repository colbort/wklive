package svc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/alert"
	"wklive/common/alert/adminnotify"
	mq "wklive/common/mq/kafka"
	"wklive/services/market/internal/config"
	"wklive/services/market/internal/market/calendar"
	"wklive/services/market/internal/market/client"
	"wklive/services/market/internal/market/types"
	"wklive/services/market/internal/pkg/itickrest"
	"wklive/services/market/internal/pkg/klinewriter"
	"wklive/services/market/internal/priceengine"
	"wklive/services/market/models"

	icache "wklive/services/market/internal/market/cache"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
	"golang.org/x/time/rate"
)

type ServiceContext struct {
	Config                       config.Config
	MarketRuntimeConfig          *config.MarketRuntimeConf
	MarketManager                *client.MarketManager
	MarketDataCache              *icache.MarketDataCache
	DataCache                    *redis.Client
	LockRedis                    *redis.Client
	TaskSubscriber               *mq.Subscriber
	SnapshotPublisher            *mq.Publisher
	OperationalAlertNotifier     alert.Notifier
	Cache                        cache.Cache
	Factory                      *models.CoinKlineModelFactory
	Writer                       *klinewriter.BatchWriter
	RebuildDerivedKlines         func([]*models.CoinKline) error
	RebuildHistoricalKlines      func([]*models.CoinKline) error
	MarketCategoryModel          models.TMarketCategoryModel
	MarketProductModel           models.TMarketProductModel
	MarketTenantCategoryModel    models.TMarketTenantCategoryModel
	MarketTenantProductModel     models.TMarketTenantProductModel
	MarketSyncTaskModel          models.TMarketSyncTaskModel
	MarketQuoteModel             models.TMarketQuoteModel
	AuthoritativeSnapshotModel   models.TMarketAuthoritativeSnapshotModel
	SnapshotOutboxModel          models.TMarketSnapshotOutboxModel
	SnapshotRevocationModel      models.TMarketSnapshotRevocationModel
	PriceFormulaModel            models.TMarketPriceFormulaModel
	PriceEngine                  *priceengine.Engine
	AuthorityRegistryModel       AuthorityRegistryStore
	AuthorityRegistryAdminModel  AuthorityRegistryAdminStore
	MarketKlineSyncProgressModel models.TMarketKlineSyncProgressModel
	MarketCalendarModel          models.TMarketMarketCalendarModel
	MarketHolidayModel           models.TMarketMarketHolidayModel
	MarketCalendarResolver       *calendar.Resolver
	ITickRestClient              *itickrest.Client
	AuthoritativeQuoteHandler    func(context.Context, types.ClientMessage, *types.QuotePayload) error
}

type AuthorityRegistryStore interface {
	FindEnabled(context.Context, string) (*models.TMarketAuthorityRegistry, error)
}

type AuthorityRegistryAdminStore interface {
	Create(context.Context, *models.TMarketAuthorityRegistry) (int64, error)
	FindOne(context.Context, int64) (*models.TMarketAuthorityRegistry, error)
	FindOneByAuthority(context.Context, string) (*models.TMarketAuthorityRegistry, error)
	FindPage(context.Context, models.AuthorityRegistryFilter, int64, int64) ([]*models.TMarketAuthorityRegistry, int64, error)
	CountActiveFormulaReferences(context.Context, string) (int64, error)
	UpdateConfigVersioned(context.Context, int64, int64, string, int64, int64) (bool, error)
}

func NewServiceContext(c config.Config) *ServiceContext {
	restRatePerMinute := c.Itick.RestRateLimitPerMinute
	if restRatePerMinute <= 0 {
		restRatePerMinute = 100
	}
	restRateBurst := c.Itick.RestRateLimitBurst
	if restRateBurst <= 0 {
		restRateBurst = 1
	}
	marketRestLimiter := rate.NewLimiter(rate.Limit(float64(restRatePerMinute)/60.0), restRateBurst)
	iTickRestClient := itickrest.New(c.Itick.Token, marketRestLimiter, nil)

	conn := sqlx.NewMysql(c.Mysql.DataSource)

	marketCategoryModel := models.NewTMarketCategoryModel(conn, c.CacheRedis)
	marketProductModel := models.NewTMarketProductModel(conn, c.CacheRedis)
	marketTenantCategoryModel := models.NewTMarketTenantCategoryModel(conn, c.CacheRedis)
	marketTenantProductModel := models.NewTMarketTenantProductModel(conn, c.CacheRedis)
	marketSyncTaskModel := models.NewTMarketSyncTaskModel(conn, c.CacheRedis)
	marketQuoteModel := models.NewTMarketQuoteModel(conn, c.CacheRedis)
	authoritativeSnapshotModel := models.NewTMarketAuthoritativeSnapshotModel(conn, c.CacheRedis)
	authorityRegistryModel := models.NewTMarketAuthorityRegistryModel(conn, c.CacheRedis)
	snapshotOutboxModel := models.NewTMarketSnapshotOutboxModel(conn, c.CacheRedis)
	snapshotRevocationModel := models.NewTMarketSnapshotRevocationModel(conn, c.CacheRedis)
	priceFormulaModel := models.NewTMarketPriceFormulaModel(conn, c.CacheRedis)
	marketKlineSyncProgressModel := models.NewTMarketKlineSyncProgressModel(conn, c.CacheRedis)
	marketCalendarModel := models.NewTMarketMarketCalendarModel(conn, c.CacheRedis)
	marketHolidayModel := models.NewTMarketMarketHolidayModel(conn, c.CacheRedis)
	marketCalendarResolver := calendar.NewResolver(marketCalendarModel, 10*time.Minute)

	dataCache := redis.NewClient(&redis.Options{
		Addr:     c.DataCache.Host,
		Username: c.DataCache.User,
		Password: c.DataCache.Pass,
		DB:       0,
	})

	lockRedis := redis.NewClient(&redis.Options{
		Addr:     c.LockRedis.Host,
		Username: c.LockRedis.User,
		Password: c.LockRedis.Pass,
		DB:       0,
	})
	marketDataCache := icache.NewMarketDataCache(dataCache)
	mqConfig := mq.ForService(c.MQ, c.Name)
	taskSubscriber := mq.MustNewSubscriber(mqConfig, "market-tasks")
	snapshotPublisher := mq.MustNewPublisher(mqConfig)

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

	marketManager := client.NewMarketManager(
		c.Itick.WSUrl,
		c.Itick.ApiUrl,
		c.Itick.Token,
		marketCategoryModel,
		marketProductModel,
		dataCache,
		lockRedis,
		marketDataCache,
		iTickRestClient,
	)
	authoritativeQuoteHandler := func(_ context.Context, msg types.ClientMessage, payload *types.QuotePayload) error {
		rpcCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := validateAuthoritativeQuoteInput(payload); err != nil {
			logx.Errorf("reject non-authoritative quote, symbol=%s market=%s err=%v", msg.Symbol, msg.Market, err)
			return err
		}
		if payload.Authority != "" && payload.LastPriceText != "" {
			authority, authorityErr := authorityRegistryModel.FindEnabled(rpcCtx, payload.Authority)
			if authorityErr != nil {
				return fmt.Errorf("resolve market authority %q: %w", payload.Authority, authorityErr)
			}
			if !authority.Allows("FINAL_QUOTE") {
				return fmt.Errorf("market authority %q cannot publish FINAL_QUOTE", payload.Authority)
			}
			snapshot, snapshotErr := icache.BuildAuthoritativeQuoteSnapshot(msg, payload)
			if snapshotErr != nil {
				logx.Errorf("publish authoritative quote failed, symbol=%s market=%s err=%v", msg.Symbol, msg.Market, snapshotErr)
				return snapshotErr
			}
			raw, marshalErr := json.Marshal(snapshot)
			outboxPayload, outboxErr := json.Marshal(map[string]any{"snapshot": snapshot, "message": msg, "quote": payload})
			price, priceErr := decimal.NewFromString(snapshot.Price)
			if marshalErr != nil || outboxErr != nil || priceErr != nil {
				logx.Errorf("encode authoritative quote failed, snapshot=%s", snapshot.SnapshotID)
				return fmt.Errorf("encode authoritative quote: snapshot=%v outbox=%v price=%v", marshalErr, outboxErr, priceErr)
			}
			if snapshotErr = authoritativeSnapshotModel.InsertImmutableAndEnqueue(rpcCtx, &models.TMarketAuthoritativeSnapshot{SnapshotId: snapshot.SnapshotID, Authority: snapshot.Authority, SnapshotKind: snapshot.Kind, CategoryCode: snapshot.CategoryCode, Market: snapshot.Market, Symbol: snapshot.Symbol, Price: price, SourceTimestamp: snapshot.SourceTimestamp, SnapshotTimestamp: snapshot.SnapshotTimestamp, Revision: snapshot.Revision, FormulaVersion: snapshot.FormulaVersion, RawPayload: string(raw), CreateTimes: time.Now().UnixMilli()}, string(outboxPayload)); snapshotErr != nil {
				logx.Errorf("archive authoritative quote failed, snapshot=%s err=%v", snapshot.SnapshotID, snapshotErr)
				return snapshotErr
			}
			return nil
		}
		return errors.New("unreachable non-authoritative quote branch")
	}
	marketManager.SetQuoteHandler(authoritativeQuoteHandler)

	return &ServiceContext{
		Config:                       c,
		MarketManager:                marketManager,
		MarketDataCache:              marketDataCache,
		DataCache:                    dataCache,
		LockRedis:                    lockRedis,
		TaskSubscriber:               taskSubscriber,
		SnapshotPublisher:            snapshotPublisher,
		OperationalAlertNotifier:     adminnotify.New(snapshotPublisher),
		Cache:                        cache.New(c.CacheRedis, syncx.NewSingleFlight(), cache.NewStat("quote"), redis.Nil),
		Factory:                      factory,
		Writer:                       writer,
		MarketCategoryModel:          marketCategoryModel,
		MarketProductModel:           marketProductModel,
		MarketTenantCategoryModel:    marketTenantCategoryModel,
		MarketTenantProductModel:     marketTenantProductModel,
		MarketSyncTaskModel:          marketSyncTaskModel,
		MarketQuoteModel:             marketQuoteModel,
		AuthoritativeSnapshotModel:   authoritativeSnapshotModel,
		SnapshotOutboxModel:          snapshotOutboxModel,
		SnapshotRevocationModel:      snapshotRevocationModel,
		PriceFormulaModel:            priceFormulaModel,
		PriceEngine:                  priceengine.New(priceFormulaModel, authoritativeSnapshotModel),
		AuthorityRegistryModel:       authorityRegistryModel,
		AuthorityRegistryAdminModel:  authorityRegistryModel,
		MarketKlineSyncProgressModel: marketKlineSyncProgressModel,
		MarketCalendarModel:          marketCalendarModel,
		MarketHolidayModel:           marketHolidayModel,
		MarketCalendarResolver:       marketCalendarResolver,
		ITickRestClient:              iTickRestClient,
		AuthoritativeQuoteHandler:    authoritativeQuoteHandler,
	}
}

func validateAuthoritativeQuoteInput(payload *types.QuotePayload) error {
	if payload == nil {
		return errors.New("quote payload is nil")
	}
	if strings.TrimSpace(payload.Authority) == "" {
		return errors.New("quote authority is required")
	}
	if strings.TrimSpace(payload.LastPriceText) == "" {
		return errors.New("quote exact decimal price is required")
	}
	return nil
}
