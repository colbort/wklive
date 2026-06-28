// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"strings"
	"wklive/app-api/internal/config"
	"wklive/app-api/internal/handler"
	"wklive/app-api/internal/svc"
	"wklive/common/etcd"
	"wklive/common/middleware"
	"wklive/common/utils"

	"github.com/zeromicro/go-zero/rest"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/app-api/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

func main() {
	flag.Parse()

	var c config.Config
	// 用 etcd 配置中心
	if err := etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c); err != nil {
		panic(err)
	}
	c.Middlewares.Log = false

	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCors("*"),
		rest.WithCorsHeaders(string(utils.CtxKeyTenantCode)),
	)
	defer server.Stop()

	requestLogMiddleware := middleware.NewRequestLogMiddleware("APP-API")
	server.Use(requestLogMiddleware.Handle)
	headerMiddleware := middleware.NewHeaderMiddleware()
	server.Use(headerMiddleware.Handle)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
