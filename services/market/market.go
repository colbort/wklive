package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	pb "wklive/proto/market"
	"wklive/services/market/internal/config"
	"wklive/services/market/internal/market/calendar"
	"wklive/services/market/internal/market/kline"
	"wklive/services/market/internal/pkg/bootstrap"
	"wklive/services/market/internal/pkg/utils"
	admin "wklive/services/market/internal/server/admin"
	app "wklive/services/market/internal/server/app"
	market "wklive/services/market/internal/server/market"
	task "wklive/services/market/internal/server/task"
	"wklive/services/market/internal/svc"
	"wklive/services/market/internal/tasks"

	"wklive/common/etcd"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/market-rpc/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

func main() {
	flag.Parse()

	var c config.Config

	// 用 etcd 配置中心
	if err := etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c); err != nil {
		panic(err)
	}

	logx.SetLevel(logx.ErrorLevel)
	mon.DisableInfoLog()

	svcCtx := svc.NewServiceContext(c)
	defer func() {
		if err := svcCtx.SnapshotPublisher.Close(); err != nil {
			log.Printf("close authoritative snapshot publisher failed: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svcCtx.MarketRuntimeConfig = loadMarketRuntimeConfig(c.Runtime)
	svcCtx.MarketDataCache.SetKlineStaleTTL(time.Duration(svcCtx.MarketRuntimeConfig.WsKlineStaleSeconds) * time.Second)
	hotWindowMinutes := c.AuthoritativeCache.HotWindowMinutes
	if hotWindowMinutes <= 0 {
		hotWindowMinutes = 30
	}
	svcCtx.MarketDataCache.SetAuthoritativeHotWindow(time.Duration(hotWindowMinutes) * time.Minute)
	tasks.StartTaskSubscriber(ctx, svcCtx)
	tasks.StartSnapshotOutbox(ctx, svcCtx)
	tasks.StartSnapshotOutboxCleanup(ctx, svcCtx)
	tasks.StartAuthoritativeSnapshotRebuild(ctx, svcCtx)
	tasks.StartLegacyAuthoritativeCacheCleanup(ctx, svcCtx)
	if err := tasks.StartExternalQuoteSources(ctx, svcCtx); err != nil {
		panic(err)
	}
	tasks.StartPriceEngine(ctx, svcCtx.PriceEngine, svcCtx.OperationalAlertNotifier)
	holidaySync := calendar.NewHolidaySyncService(ctx, c.Itick.ApiUrl, svcCtx.ITickRestClient,
		svcCtx.MarketCalendarModel, svcCtx.MarketHolidayModel, svcCtx.MarketCalendarResolver,
		utils.NewRedisLock(svcCtx.LockRedis), 24*time.Hour)
	holidaySync.Start()
	defer holidaySync.Stop()

	// 预热的 categoryCode + interval，自行按你的业务改
	if err := bootstrap.PreheatCoinKlineModels(svcCtx.Factory); err != nil {
		log.Fatalf("preheat coin kline models failed: %v", err)
	}

	// 启动批量写入器
	svcCtx.Writer.Start()
	derivedAggregator := kline.NewDerivedAggregator(svcCtx.Factory, svcCtx.MarketDataCache, svcCtx.MarketCalendarResolver)
	derivedWorker := kline.NewDerivedWorker(derivedAggregator, 1024)
	derivedWorker.Start()
	defer derivedWorker.Stop()
	defer svcCtx.Writer.Stop()
	svcCtx.RebuildDerivedKlines = derivedWorker.Rebuild
	svcCtx.RebuildHistoricalKlines = derivedWorker.RebuildHistory
	gapRepair := kline.NewGapRepairService(ctx, svcCtx,
		time.Duration(svcCtx.MarketRuntimeConfig.GapScanIntervalMinutes)*time.Minute,
		int(svcCtx.MarketRuntimeConfig.RepairBatchSize))
	gapRepair.Start(c.Itick.ApiUrl, c.Itick.Token)
	defer gapRepair.Stop()
	svcCtx.MarketManager.SetReconnectHandler(func(providerCode, category string) {
		if providerCode != "itick" {
			log.Printf("skip iTick kline repair after non-iTick reconnect, provider=%s category=%s", providerCode, category)
			return
		}
		repairCtx, repairCancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer repairCancel()
		worker := kline.NewSyncKlinesWorker(repairCtx, svcCtx, nil, "", "")
		if err := worker.RepairAfterReconnect(c.Itick.ApiUrl, c.Itick.Token, category); err != nil {
			log.Printf("repair market klines after ws reconnect failed, provider=%s category=%s err=%v", providerCode, category, err)
		}
		if err := worker.ReconcileRecent(c.Itick.ApiUrl, c.Itick.Token, category); err != nil {
			log.Printf("reconcile recent market klines after ws reconnect failed, provider=%s category=%s err=%v", providerCode, category, err)
		}
	})
	svcCtx.Writer.SetFlushHandler(derivedWorker.Enqueue)
	tickAggregator := kline.NewTickAggregator(
		svcCtx.Writer,
		svcCtx.DataCache,
		time.Duration(svcCtx.MarketRuntimeConfig.BuildingBucketTtlMinutes)*time.Minute,
	)
	svcCtx.MarketDataCache.SetTickHandler(tickAggregator.Add)
	tickAggregator.Start()
	defer tickAggregator.Stop()

	// 加载 market 分类数据并初始化 WebSocket 客户端
	err := svcCtx.MarketManager.Load(ctx)
	if err != nil {
		panic(err)
	}
	if err := svcCtx.MarketManager.LoadActiveProductSubscriptions(ctx); err != nil {
		panic(err)
	}
	// 启动 market 数据流管理器
	if err := svcCtx.MarketManager.Start(ctx); err != nil {
		panic(err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterAdminServer(grpcServer, admin.NewAdminServer(svcCtx))
		pb.RegisterAppServer(grpcServer, app.NewAppServer(svcCtx))
		pb.RegisterMarketServer(grpcServer, market.NewMarketServer(svcCtx))
		pb.RegisterTaskServer(grpcServer, task.NewTaskServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

func loadMarketRuntimeConfig(runtime config.MarketRuntimeConf) *config.MarketRuntimeConf {
	if runtime.ReconcileIntervalMinutes <= 0 {
		runtime.ReconcileIntervalMinutes = 5
	}
	if runtime.ReconcileWindowBars <= 0 {
		runtime.ReconcileWindowBars = 30
	}
	if runtime.GapScanIntervalMinutes <= 0 {
		runtime.GapScanIntervalMinutes = 60
	}
	if runtime.RepairBatchSize <= 0 {
		runtime.RepairBatchSize = 10
	}
	if runtime.BuildingBucketTtlMinutes <= 0 {
		runtime.BuildingBucketTtlMinutes = 120
	}
	if runtime.WsKlineStaleSeconds <= 0 {
		runtime.WsKlineStaleSeconds = 30
	}
	return &runtime
}
