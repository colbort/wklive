// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"strconv"
	"wklive/admin-api/internal/config"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/itick"
	"wklive/proto/option"
	"wklive/proto/payment"
	"wklive/proto/staking"
	"wklive/proto/system"
	"wklive/proto/trade"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServiceContext struct {
	Config     config.Config
	SystemCli  system.SystemClient
	UserCli    user.UserAdminClient
	PaymentCli payment.PaymentAdminClient
	ItickCli   itick.ItickAdminClient
	AssetCli   asset.AssetAdminClient
	OptionCli  option.OptionAdminClient
	StakingCli staking.StakingAdminClient
	TradeCli   trade.TradeAdminClient
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
	systemCli := zrpc.MustNewClient(c.SystemRpc, options)
	userCli := zrpc.MustNewClient(c.UserRpc, options)
	paymentCli := zrpc.MustNewClient(c.PaymentRpc, options)
	itickCli := zrpc.MustNewClient(c.ItickRpc, options)
	assetCli := zrpc.MustNewClient(c.AssetRpc, options)
	optionCli := zrpc.MustNewClient(c.OptionRpc, options)
	stakingCli := zrpc.MustNewClient(c.StakingRpc, options)
	tradeCli := zrpc.MustNewClient(c.TradeRpc, options)
	return &ServiceContext{
		Config:     c,
		SystemCli:  system.NewSystemClient(systemCli.Conn()),
		UserCli:    user.NewUserAdminClient(userCli.Conn()),
		PaymentCli: payment.NewPaymentAdminClient(paymentCli.Conn()),
		ItickCli:   itick.NewItickAdminClient(itickCli.Conn()),
		AssetCli:   asset.NewAssetAdminClient(assetCli.Conn()),
		OptionCli:  option.NewOptionAdminClient(optionCli.Conn()),
		StakingCli: staking.NewStakingAdminClient(stakingCli.Conn()),
		TradeCli:   trade.NewTradeAdminClient(tradeCli.Conn()),
	}
}
