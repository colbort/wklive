// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"

	"chat-api/internal/config"
	"chat-api/internal/middleware"
	"chat-api/internal/ws"
	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	UserRateLimit  rest.Middleware
	ChatAppCli     chat.ChatAppClient
	ChatMessageHub *ws.Hub
}

func NewServiceContext(c config.Config) *ServiceContext {
	chatCli := zrpc.MustNewClient(c.ChatRpc)
	chatMessageHub := ws.NewHub()
	go chatMessageHub.Run()
	if c.RedisConf.Host != "" {
		go ws.SubscribeRedis(context.Background(), c.RedisConf, chatMessageHub)
	} else {
		logx.Info("chat api redis is not configured, skip message subscription")
	}

	return &ServiceContext{
		Config:         c,
		UserRateLimit:  middleware.NewUserRateLimitMiddleware().Handle,
		ChatAppCli:     chat.NewChatAppClient(chatCli.Conn()),
		ChatMessageHub: chatMessageHub,
	}
}
