// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"chat-api/internal/config"
	"chat-api/internal/handler"
	"chat-api/internal/svc"
	"wklive/common/etcd"
	"wklive/common/middleware"

	"github.com/zeromicro/go-zero/rest"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/chat-api/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

func main() {
	flag.Parse()

	var c config.Config
	if err := etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c); err != nil {
		panic(err)
	}
	c.Middlewares.Log = false

	var opts []rest.RunOption
	if len(c.Cors) > 0 {
		opts = append(opts, rest.WithCors(c.Cors...), rest.WithFileServer(
			"/chat_uploads",
			http.Dir("./chat_uploads"),
		))
	}
	server := rest.MustNewServer(c.RestConf, opts...)
	defer server.Stop()

	requestLogMiddleware := middleware.NewRequestLogMiddleware("CHAT-API")
	server.Use(requestLogMiddleware.Handle)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	handler.RegisterCustomHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
