// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"strconv"
	"wklive/app-api/internal/config"
	"wklive/app-api/internal/middleware"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/itick"
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
	SystemCli               system.SystemClient
	UserCli                 user.UserAppClient
	PaymentCli              payment.PaymentAppClient
	ItickCli                itick.ItickAppClient
	AssetCli                asset.AssetAppClient
	OptionCli               option.OptionAppClient
	StakingCli              staking.StakingAppClient
	TradeCli                trade.TradeAppClient
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
		pairs := make([]string, 0, 4)
		if uid, err := utils.GetUidFromCtx(ctx); err == nil {
			pairs = append(pairs, "uid", strconv.FormatInt(uid, 10))
		}
		if username, err := utils.GetUsernameFromCtx(ctx); err == nil {
			pairs = append(pairs, "username", username)
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
	itickCli := zrpc.MustNewClient(c.ItickRpc, options)
	assetCli := zrpc.MustNewClient(c.AssetRpc, options)
	optionCli := zrpc.MustNewClient(c.OptionRpc, options)
	stakingCli := zrpc.MustNewClient(c.StakingRpc, options)
	tradeCli := zrpc.MustNewClient(c.TradeRpc, options)
	return &ServiceContext{
		Config:                  c,
		PublicRateLimit:         middleware.NewPublicRateLimitMiddleware(rds).Handle,
		GuestSensitiveRateLimit: middleware.NewGuestSensitiveRateLimitMiddleware(rds).Handle,
		UserRateLimit:           middleware.NewUserRateLimitMiddleware(rds).Handle,
		SensitiveRateLimit:      middleware.NewSensitiveRateLimitMiddleware(rds).Handle,
		RefreshTokenRateLimit:   middleware.NewRefreshTokenRateLimitMiddleware(rds).Handle,
		SystemCli:               system.NewSystemClient(systemCli.Conn()),
		UserCli:                 user.NewUserAppClient(userCli.Conn()),
		PaymentCli:              payment.NewPaymentAppClient(paymentCli.Conn()),
		ItickCli:                itick.NewItickAppClient(itickCli.Conn()),
		AssetCli:                asset.NewAssetAppClient(assetCli.Conn()),
		OptionCli:               option.NewOptionAppClient(optionCli.Conn()),
		StakingCli:              staking.NewStakingAppClient(stakingCli.Conn()),
		TradeCli:                trade.NewTradeAppClient(tradeCli.Conn()),
	}
}
