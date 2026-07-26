package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	pb "wklive/proto/itick"
	"wklive/proto/system"
	"wklive/services/itick/internal/config"
	"wklive/services/itick/internal/market/calendar"
	"wklive/services/itick/internal/market/kline"
	"wklive/services/itick/internal/pkg/bootstrap"
	"wklive/services/itick/internal/pkg/utils"
	admin "wklive/services/itick/internal/server/admin"
	app "wklive/services/itick/internal/server/app"
	itick "wklive/services/itick/internal/server/itick"
	task "wklive/services/itick/internal/server/task"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/internal/tasks"

	"wklive/common/etcd"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/itick-rpc/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

func main() {
	flag.Parse()

	var c config.Config

	// 用 etcd 配置中心
	if err := etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c); err != nil {
		panic(err)
	}

	svcCtx := svc.NewServiceContext(c)
	defer func() {
		if err := svcCtx.SnapshotPublisher.Close(); err != nil {
			log.Printf("close authoritative snapshot publisher failed: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svcCtx.ItickRuntimeConfig = loadItickRuntimeConfig(ctx, svcCtx.SystemCli)
	svcCtx.MarketDataCache.SetKlineStaleTTL(time.Duration(svcCtx.ItickRuntimeConfig.WsKlineStaleSeconds) * time.Second)
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
	tasks.StartPriceEngine(ctx, svcCtx.PriceEngine)
	holidaySync := calendar.NewHolidaySyncService(ctx, c.Itick.ApiUrl, svcCtx.ItickRestClient,
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
		time.Duration(svcCtx.ItickRuntimeConfig.GapScanIntervalMinutes)*time.Minute,
		int(svcCtx.ItickRuntimeConfig.RepairBatchSize))
	gapRepair.Start(c.Itick.ApiUrl, c.Itick.Token)
	defer gapRepair.Stop()
	svcCtx.ItickManager.SetReconnectHandler(func(category string) {
		repairCtx, repairCancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer repairCancel()
		worker := kline.NewSyncKlinesWorker(repairCtx, svcCtx, nil, "", "")
		if err := worker.RepairAfterReconnect(c.Itick.ApiUrl, c.Itick.Token, category); err != nil {
			log.Printf("repair itick klines after ws reconnect failed, category=%s err=%v", category, err)
		}
		if err := worker.ReconcileRecent(c.Itick.ApiUrl, c.Itick.Token, category); err != nil {
			log.Printf("reconcile recent itick klines after ws reconnect failed, category=%s err=%v", category, err)
		}
	})
	svcCtx.Writer.SetFlushHandler(derivedWorker.Enqueue)
	tickAggregator := kline.NewTickAggregator(
		svcCtx.Writer,
		svcCtx.DataCache,
		time.Duration(svcCtx.ItickRuntimeConfig.BuildingBucketTtlMinutes)*time.Minute,
	)
	svcCtx.MarketDataCache.SetTickHandler(tickAggregator.Add)
	tickAggregator.Start()
	defer tickAggregator.Stop()

	// 加载 itick 分类数据并初始化 WebSocket 客户端
	err := svcCtx.ItickManager.Load(ctx)
	if err != nil {
		panic(err)
	}
	if err := svcCtx.ItickManager.LoadActiveProductSubscriptions(ctx); err != nil {
		panic(err)
	}
	// 启动 itick 数据流管理器
	if err := svcCtx.ItickManager.Start(ctx); err != nil {
		panic(err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterAdminServer(grpcServer, admin.NewAdminServer(svcCtx))
		pb.RegisterAppServer(grpcServer, app.NewAppServer(svcCtx))
		pb.RegisterItickServer(grpcServer, itick.NewItickServer(svcCtx))
		pb.RegisterTaskServer(grpcServer, task.NewTaskServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

func loadItickRuntimeConfig(ctx context.Context, cli system.SystemClient) *system.ItickConfig {
	config := &system.ItickConfig{ReconcileIntervalMinutes: 5, ReconcileWindowBars: 30,
		GapScanIntervalMinutes: 60, RepairBatchSize: 10, BuildingBucketTtlMinutes: 120, WsKlineStaleSeconds: 30}
	key := system.SysConfigType_ITICK_CONFIG
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	resp, err := cli.SysConfigDetail(callCtx, &system.SysConfigDetailReq{ConfigKey: &key})
	if err == nil && resp.GetData() != nil {
		_ = json.Unmarshal([]byte(resp.GetData().GetConfigValue()), config)
	}
	if config.ReconcileIntervalMinutes <= 0 {
		config.ReconcileIntervalMinutes = 5
	}
	if config.ReconcileWindowBars <= 0 {
		config.ReconcileWindowBars = 30
	}
	if config.GapScanIntervalMinutes <= 0 {
		config.GapScanIntervalMinutes = 60
	}
	if config.RepairBatchSize <= 0 {
		config.RepairBatchSize = 10
	}
	if config.BuildingBucketTtlMinutes <= 0 {
		config.BuildingBucketTtlMinutes = 120
	}
	if config.WsKlineStaleSeconds <= 0 {
		config.WsKlineStaleSeconds = 30
	}
	return config
}
