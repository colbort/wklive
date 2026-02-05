package main

import (
	"flag"
	"fmt"
	"strings"

	"wklive/rpc/system"
	"wklive/services/system/internal/config"
	"wklive/services/system/internal/server"
	"wklive/services/system/internal/svc"

	"wklive/common/etcd"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "192.168.10.116:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/system-rpc/config", "etcd config key")
)

func main() {
	flag.Parse()

	var c config.Config

	// 用 etcd 配置中心
	etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), *configKey, &c)

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		system.RegisterSystemServer(grpcServer, server.NewSystemServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
