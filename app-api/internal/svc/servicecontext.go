// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"wklive/app-api/internal/config"
	"wklive/app-api/internal/middleware"
	apprealtime "wklive/app-api/internal/realtime"
	mq "wklive/common/mq/kafka"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/market"
	"wklive/proto/option"
	"wklive/proto/payment"
	"wklive/proto/staking"
	"wklive/proto/system"
	"wklive/proto/trade"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServiceContext struct {
	Config                  config.Config
	PublicRateLimit         rest.Middleware
	GuestSensitiveRateLimit rest.Middleware
	UserRateLimit           rest.Middleware
	SensitiveRateLimit      rest.Middleware
	RefreshTokenRateLimit   rest.Middleware
	SystemCli               system.AppClient
	UserCli                 user.AppClient
	PaymentCli              payment.AppClient
	MarketCli               market.AppClient
	AssetCli                asset.AppClient
	OptionCli               option.AppClient
	StakingCli              staking.AppClient
	TradeCli                trade.AppClient
	UserEventHub            *apprealtime.UserEventHub
	UserEventSubscriber     *mq.Subscriber
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
		pairs := make([]string, 0, 5)
		if userId, err := utils.GetUserIdFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyUid, strconv.FormatInt(userId, 10))
		}
		if username, err := utils.GetUsernameFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyUsername, username)
		}
		if tenantId, err := utils.GetTrustedTenantIdFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyTenantId, fmt.Sprintf("%d", tenantId))
		}
		if tenantCode, err := utils.GetTenantCodeFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyTenantCode, tenantCode)
		}
		if clientIP, err := utils.GetClientIPFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyClientIp, clientIP)
		}
		if len(pairs) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	})
	rds := redis.MustNewRedis(c.RedisConf)
	systemCli := zrpc.MustNewClient(c.SystemRpc, options)
	userCli := zrpc.MustNewClient(c.UserRpc, options)
	paymentCli := zrpc.MustNewClient(c.PaymentRpc, options)
	marketCli := zrpc.MustNewClient(c.MarketRpc, options)
	assetCli := zrpc.MustNewClient(c.AssetRpc, options)
	optionCli := zrpc.MustNewClient(c.OptionRpc, options)
	stakingCli := zrpc.MustNewClient(c.StakingRpc, options)
	tradeCli := zrpc.MustNewClient(c.TradeRpc, options)
	instanceID, err := os.Hostname()
	if err != nil || instanceID == "" {
		instanceID = "local"
	}
	userEventHub := apprealtime.NewUserEventHub()
	userEventSubscriber := mq.MustNewBroadcastSubscriber(
		mq.ForService(c.MQ, c.Name),
		"app-api-user-events-"+instanceID,
	)
	return &ServiceContext{
		Config:                  c,
		PublicRateLimit:         middleware.NewPublicRateLimitMiddleware(rds).Handle,
		GuestSensitiveRateLimit: middleware.NewGuestSensitiveRateLimitMiddleware(rds).Handle,
		UserRateLimit:           middleware.NewUserRateLimitMiddleware(rds).Handle,
		SensitiveRateLimit:      middleware.NewSensitiveRateLimitMiddleware(rds).Handle,
		RefreshTokenRateLimit:   middleware.NewRefreshTokenRateLimitMiddleware(rds).Handle,
		SystemCli:               system.NewAppClient(systemCli.Conn()),
		UserCli:                 user.NewAppClient(userCli.Conn()),
		PaymentCli:              payment.NewAppClient(paymentCli.Conn()),
		MarketCli:               market.NewAppClient(marketCli.Conn()),
		AssetCli:                asset.NewAppClient(assetCli.Conn()),
		OptionCli:               option.NewAppClient(optionCli.Conn()),
		StakingCli:              staking.NewAppClient(stakingCli.Conn()),
		TradeCli:                trade.NewAppClient(tradeCli.Conn()),
		UserEventHub:            userEventHub,
		UserEventSubscriber:     userEventSubscriber,
	}
}
