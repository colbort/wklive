package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"wklive/proto/itick"
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
	endpoints = flag.String("etcd", "192.168.100.116:2379", "etcd endpoints")
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

	// 启动 itick 数据流管理器
	svcCtx.ItickManager.Start(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		itick.RegisterItickAdminServer(grpcServer, server.NewItickAdminServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
