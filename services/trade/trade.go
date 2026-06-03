package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"wklive/proto/trade"
	"wklive/services/trade/internal/config"
	"wklive/services/trade/internal/logic"
	"wklive/services/trade/internal/server"
	"wklive/services/trade/internal/svc"

	"wklive/common/etcd"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/trade-rpc/config", "etcd config key")
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
	if restored, err := logic.RestoreOrderBookCache(context.Background(), svcCtx); err != nil {
		fmt.Printf("Restore order book cache failed: %v\n", err)
	} else {
		fmt.Printf("Restored %d open orders into order book cache.\n", restored)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		trade.RegisterTradeAdminServer(grpcServer, server.NewTradeAdminServer(svcCtx))
		trade.RegisterTradeAppServer(grpcServer, server.NewTradeAppServer(svcCtx))
		trade.RegisterTradeInternalServer(grpcServer, server.NewTradeInternalServer(svcCtx))
		trade.RegisterTradeTaskServer(grpcServer, server.NewTradeTaskServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
