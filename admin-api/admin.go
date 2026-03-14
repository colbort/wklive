// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
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

	// 添加静态文件路由，提供头像访问
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/avatars/*filepath",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			// 提取文件名（去掉 /avatars/ 前缀）
			fname := strings.TrimPrefix(r.URL.Path, "/avatars/")
			// 构建本地路径
			path := filepath.Join("/var/www/avatars", fname)
			// 服务文件
			http.ServeFile(w, r, path)
		},
	})

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
