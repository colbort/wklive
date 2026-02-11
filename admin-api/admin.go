// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"strings"

	"wklive/admin-api/internal/config"
	"wklive/admin-api/internal/handler"
	"wklive/admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"

	"wklive/common/etcd"
)

var (
	endpoints = flag.String("etcd", "192.168.10.116:2379", "etcd endpoints")
	configKey = flag.String("config", "/wklive/admin-api/config", "etcd config key")
	commonKey = flag.String("common", "/wklive/common/config", "etcd common config key")
)

func main() {
	flag.Parse()

	var c config.Config
	// 用 etcd 配置中心
	etcd.LoadFromEtcdAndMerge(strings.Split(*endpoints, ","), []string{*commonKey, *configKey}, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
