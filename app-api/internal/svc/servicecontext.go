// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"wklive/app-api/internal/config"
	"wklive/proto/asset"
	"wklive/proto/itick"
	"wklive/proto/option"
	"wklive/proto/payment"
	"wklive/proto/staking"
	"wklive/proto/system"
	"wklive/proto/trade"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	SystemCli  system.SystemClient
	UserCli    user.UserAppClient
	PaymentCli payment.PaymentAppClient
	ItickCli   itick.ItickAppClient
	AssetCli   asset.AssetAppClient
	OptionCli  option.OptionAppClient
	StakingCli staking.StakingAppClient
	TradeCli   trade.TradeAppClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	systemCli := zrpc.MustNewClient(c.SystemRpc)
	userCli := zrpc.MustNewClient(c.UserRpc)
	paymentCli := zrpc.MustNewClient(c.PaymentRpc)
	itickCli := zrpc.MustNewClient(c.ItickRpc)
	assetCli := zrpc.MustNewClient(c.AssetRpc)
	optionCli := zrpc.MustNewClient(c.OptionRpc)
	stakingCli := zrpc.MustNewClient(c.StakingRpc)
	tradeCli := zrpc.MustNewClient(c.TradeRpc)
	return &ServiceContext{
		Config:     c,
		SystemCli:  system.NewSystemClient(systemCli.Conn()),
		UserCli:    user.NewUserAppClient(userCli.Conn()),
		PaymentCli: payment.NewPaymentAppClient(paymentCli.Conn()),
		ItickCli:   itick.NewItickAppClient(itickCli.Conn()),
		AssetCli:   asset.NewAssetAppClient(assetCli.Conn()),
		OptionCli:  option.NewOptionAppClient(optionCli.Conn()),
		StakingCli: staking.NewStakingAppClient(stakingCli.Conn()),
		TradeCli:   trade.NewTradeAppClient(tradeCli.Conn()),
	}
}
