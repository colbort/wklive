// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"

	"chat-admin-api/internal/config"
	"chat-admin-api/internal/middleware"
	"chat-admin-api/internal/ws"
	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	AdminRateLimit rest.Middleware
	ChatAdminCli   chat.ChatAdminClient
	ChatMessageHub *ws.Hub
	BusRedis       *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	chatCli := zrpc.MustNewClient(c.ChatRpc)
	chatMessageHub := ws.NewHub()
	var chatBusRedis *redis.Redis
	go chatMessageHub.Run()
	if c.RedisConf.Host != "" {
		rds, err := redis.NewRedis(c.RedisConf)
		if err != nil {
			logx.Errorf("chat admin redis init failed: %v", err)
		} else {
			chatBusRedis = rds
		}
		go ws.SubscribeRedis(context.Background(), c.RedisConf, chatMessageHub)
	} else {
		logx.Info("chat admin redis is not configured, skip message subscription")
	}

	return &ServiceContext{
		Config:         c,
		AdminRateLimit: middleware.NewAdminRateLimitMiddleware().Handle,
		ChatAdminCli:   chat.NewChatAdminClient(chatCli.Conn()),
		ChatMessageHub: chatMessageHub,
		BusRedis:       chatBusRedis,
	}
}
