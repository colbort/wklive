package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	pb "wklive/proto/payment"
	"wklive/services/payment/internal/config"
	"wklive/services/payment/internal/outbox"
	admin "wklive/services/payment/internal/server/admin"
	app "wklive/services/payment/internal/server/app"
	callback "wklive/services/payment/internal/server/callback"
	"wklive/services/payment/internal/svc"

	"wklive/common/etcd"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/payment-rpc/config", "etcd config key")
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
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	outbox.Start(runCtx, ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterAdminServer(grpcServer, admin.NewAdminServer(ctx))
		pb.RegisterAppServer(grpcServer, app.NewAppServer(ctx))
		pb.RegisterCallbackServer(grpcServer, callback.NewCallbackServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
