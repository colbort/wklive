package main

import (
	"flag"
	"fmt"
	"strings"

	pb "wklive/proto/user"
	"wklive/services/user/internal/config"
	admin "wklive/services/user/internal/server/admin"
	app "wklive/services/user/internal/server/app"
	user "wklive/services/user/internal/server/user"
	"wklive/services/user/internal/svc"

	"wklive/common/etcd"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/user-rpc/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

type watcherConfig struct {
	GuestTransfer struct {
		ExpireSeconds int64
	} `json:"GuestTransfer" yaml:"GuestTransfer"`
}

func main() {
	flag.Parse()

	var c config.Config

	// 用 etcd 配置中心
	if err := etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c); err != nil {
		panic(err)
	}

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterAdminServer(grpcServer, admin.NewAdminServer(ctx))
		pb.RegisterAppServer(grpcServer, app.NewAppServer(ctx))
		pb.RegisterUserServer(grpcServer, user.NewUserServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	etcd.WatcherConfig(strings.Split(*endpoints, ","), *configKey, func(v watcherConfig) {
		ctx.Config.GuestTransfer = v.GuestTransfer
	})

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
