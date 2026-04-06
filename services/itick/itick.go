package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"wklive/proto/itick"
	"wklive/services/itick/internal/bootstrap"
	"wklive/services/itick/internal/config"
	"wklive/services/itick/internal/server"
	"wklive/services/itick/internal/svc"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 预热的 categoryCode + interval，自行按你的业务改
	if err := bootstrap.PreheatCoinKlineModels(svcCtx.Factory); err != nil {
		log.Fatalf("preheat coin kline models failed: %v", err)
	}

	// 启动批量写入器
	svcCtx.Writer.Start()
	defer svcCtx.Writer.Stop()

	// 加载 itick 分类数据并初始化 WebSocket 客户端
	err := svcCtx.ItickManager.Load(ctx)
	if err != nil {
		panic(err)
	}
	// 启动 itick 数据流管理器
	svcCtx.ItickManager.Start(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		itick.RegisterItickAdminServer(grpcServer, server.NewItickAdminServer(svcCtx))
		itick.RegisterItickAppServer(grpcServer, server.NewItickAppServer(svcCtx))
		itick.RegisterItickTaskServer(grpcServer, server.NewItickTaskServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
