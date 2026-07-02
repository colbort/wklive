// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"chat-admin-api/internal/config"
	"chat-admin-api/internal/handler"
	"chat-admin-api/internal/svc"
	"wklive/common/etcd"
	"wklive/common/middleware"
	"wklive/common/utils"

	"github.com/zeromicro/go-zero/rest"
)

var (
	endpoints = flag.String("etcd", "127.0.0.1:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/chat-admin-api/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

func main() {
	flag.Parse()

	var c config.Config
	if err := etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c); err != nil {
		panic(err)
	}
	c.Middlewares.Log = false

	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCors("*"),
		rest.WithCorsHeaders(string(utils.CtxKeyMerchantId)),
		rest.WithCorsHeaders(string(utils.CtxKeyChatUserId)),
		rest.WithFileServer(
			"/chat_uploads",
			http.Dir(c.ChatUploadDir),
		),
	)
	defer server.Stop()

	requestLogMiddleware := middleware.NewRequestLogMiddleware("CHAT-ADMIN-API")
	server.Use(requestLogMiddleware.Handle)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	handler.RegisterCustomHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
