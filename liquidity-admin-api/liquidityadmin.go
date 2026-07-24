// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"strings"

	"wklive/liquidity-admin-api/internal/config"
	"wklive/liquidity-admin-api/internal/handler"
	"wklive/liquidity-admin-api/internal/svc"

	"wklive/common/etcd"
	um "wklive/common/middleware"
	"wklive/liquidity-admin-api/internal/middleware"

	"github.com/zeromicro/go-zero/rest"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/liquidity-admin-api/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

func main() {
	flag.Parse()

	var c config.Config
	if err := etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c); err != nil {
		panic(err)
	}

	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCors("*"),
	)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	server.Use(um.NewRequestLogMiddleware("LIQUIDITY-ADMIN-API").Handle)
	server.Use(middleware.NewRbacMiddleware(ctx).Handle)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
