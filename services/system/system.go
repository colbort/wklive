package main

import (
	"flag"
	"fmt"
	"strings"

	"wklive/proto/system"
	"wklive/services/system/internal/bootstrap"
	"wklive/services/system/internal/config"
	"wklive/services/system/internal/server"
	"wklive/services/system/internal/svc"

	"wklive/common/etcd"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/system-rpc/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

func main() {
	flag.Parse()

	var c config.Config

	// 用 etcd 配置中心
	if err := etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c); err != nil {
		panic(err)
	}

	ctx := svc.NewServiceContext(c)

	// 加载定时任务
	if err := bootstrap.LoadJobs(ctx); err != nil {
		logx.Errorf("load cron jobs failed: %v", err)
	}

	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		system.RegisterSystemServer(grpcServer, server.NewSystemServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	defer server.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	server.Start()
}
