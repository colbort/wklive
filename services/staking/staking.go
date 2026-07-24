package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	pb "wklive/proto/staking"
	"wklive/services/staking/internal/config"
	logic "wklive/services/staking/internal/logic/task"
	admin "wklive/services/staking/internal/server/admin"
	app "wklive/services/staking/internal/server/app"
	task "wklive/services/staking/internal/server/task"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/internal/tasks"

	"wklive/common/etcd"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/staking-rpc/config", "etcd config key")
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
	tasks.StartTaskSubscriber(ctx, svcCtx)
	logic.StartDelayQueue(ctx, svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterAdminServer(grpcServer, admin.NewAdminServer(svcCtx))
		pb.RegisterAppServer(grpcServer, app.NewAppServer(svcCtx))
		pb.RegisterTaskServer(grpcServer, task.NewTaskServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
