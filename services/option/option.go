package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"wklive/proto/option"
	"wklive/services/option/internal/config"
	admin "wklive/services/option/internal/server/admin"
	app "wklive/services/option/internal/server/app"
	options "wklive/services/option/internal/server/option"
	task "wklive/services/option/internal/server/task"
	"wklive/services/option/internal/svc"
	"wklive/services/option/internal/tasks"

	"wklive/common/etcd"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/option-rpc/config", "etcd config key")
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

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		option.RegisterAdminServer(grpcServer, admin.NewAdminServer(svcCtx))
		option.RegisterAppServer(grpcServer, app.NewAppServer(svcCtx))
		option.RegisterOptionServer(grpcServer, options.NewOptionServer(svcCtx))
		option.RegisterTaskServer(grpcServer, task.NewTaskServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
