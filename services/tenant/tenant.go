package main

import (
	"flag"
	"fmt"
	"strings"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/config"
	"wklive/services/tenant/internal/server"
	"wklive/services/tenant/internal/svc"

	"wklive/common/etcd"

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

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		tenant.RegisterTenantServer(grpcServer, server.NewTenantServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
