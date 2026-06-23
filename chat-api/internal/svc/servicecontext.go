// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"strconv"

	"chat-api/internal/config"
	"chat-api/internal/middleware"
	"chat-api/internal/ws"
	common "wklive/common/middleware"
	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServiceContext struct {
	Config         config.Config
	UserRateLimit  rest.Middleware
	HeaderIdentity rest.Middleware
	ChatAppCli     chat.ChatAppClient
	ChatMessageHub *ws.Hub
	BusRedis       *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	options := zrpc.WithUnaryClientInterceptor(func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		pairs := make([]string, 0, 2)
		if userID, err := utils.GetUserIdFromCtx(ctx); err == nil && userID > 0 {
			pairs = append(pairs, utils.CtxKeyUid, strconv.FormatInt(userID, 10))
		}
		if username, err := utils.GetUsernameFromCtx(ctx); err == nil && username != "" {
			pairs = append(pairs, utils.CtxKeyUsername, username)
		}
		if merchantId, err := utils.GetMerchantIdFromCtx(ctx); err == nil && merchantId > 0 {
			pairs = append(pairs, utils.CtxKeyMerchantId, strconv.FormatInt(merchantId, 10))
		}
		if len(pairs) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	})
	chatCli := zrpc.MustNewClient(c.ChatRpc, options)
	chatMessageHub := ws.NewHub()
	var chatBusRedis *redis.Redis
	go chatMessageHub.Run()
	if c.RedisConf.Host != "" {
		rds, err := redis.NewRedis(c.RedisConf)
		if err != nil {
			logx.Errorf("chat api redis init failed: %v", err)
		} else {
			chatBusRedis = rds
		}
		go ws.SubscribeRedis(context.Background(), c.RedisConf, chatMessageHub)
	} else {
		logx.Info("chat api redis is not configured, skip message subscription")
	}

	return &ServiceContext{
		Config:         c,
		UserRateLimit:  middleware.NewUserRateLimitMiddleware().Handle,
		HeaderIdentity: common.NewHeaderMiddleware().Handle,
		ChatAppCli:     chat.NewChatAppClient(chatCli.Conn()),
		ChatMessageHub: chatMessageHub,
		BusRedis:       chatBusRedis,
	}
}
