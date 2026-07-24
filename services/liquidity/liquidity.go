package main

import (
	"flag"
	"fmt"
	"strings"

	"wklive/common/etcd"
	pb "wklive/proto/liquidity"
	"wklive/services/liquidity/internal/config"
	admin "wklive/services/liquidity/internal/server/admin"
	liquidity "wklive/services/liquidity/internal/server/liquidity"
	task "wklive/services/liquidity/internal/server/task"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/liquidity-rpc/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

func main() {
	flag.Parse()

	var c config.Config
	if err := etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c); err != nil {
		panic(err)
	}

	svcCtx := svc.NewServiceContext(c)
	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterAdminServer(grpcServer, admin.NewAdminServer(svcCtx))
		pb.RegisterLiquidityServer(grpcServer, liquidity.NewLiquidityServer(svcCtx))
		pb.RegisterTaskServer(grpcServer, task.NewTaskServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer server.Stop()

	fmt.Printf("Starting liquidity rpc server at %s...\n", c.ListenOn)
	server.Start()
}
