// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"wklive/admin-api/internal/config"
	"wklive/admin-api/internal/handler"
	"wklive/admin-api/internal/middleware"
	"wklive/admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"

	"wklive/common/etcd"
	um "wklive/common/middleware"
	"wklive/common/utils"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/admin-api/config", "etcd config key")
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
		rest.WithCorsHeaders(string(utils.CtxKeyTenantId)),
		rest.WithFileServer(
			"/avatars",
			http.Dir("./avatars"),
		),
	)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	requestLogMiddleware := um.NewRequestLogMiddleware("ADMIN-API")
	server.Use(requestLogMiddleware.Handle)
	headerMiddleware := um.NewHeaderMiddleware()
	server.Use(headerMiddleware.Handle)
	rbacMiddleware := middleware.NewRbacMiddleware(ctx)
	server.Use(rbacMiddleware.Handle)

	handler.RegisterHandlers(server, ctx)
	handler.RegisterCustomHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
