// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"wklive/app-api/internal/config"
	"wklive/proto/itick"
	"wklive/proto/payment"
	"wklive/proto/system"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	SystemCli  system.SystemClient
	UserCli    user.UserAppClient
	PaymentCli payment.PaymentAppClient
	ItickCli   itick.ItickAppClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	systemCli := zrpc.MustNewClient(c.SystemRpc)
	userCli := zrpc.MustNewClient(c.UserRpc)
	paymentCli := zrpc.MustNewClient(c.PaymentRpc)
	itickCli := zrpc.MustNewClient(c.ItickRpc)
	return &ServiceContext{
		Config:     c,
		SystemCli:  system.NewSystemClient(systemCli.Conn()),
		UserCli:    user.NewUserAppClient(userCli.Conn()),
		PaymentCli: payment.NewPaymentAppClient(paymentCli.Conn()),
		ItickCli:   itick.NewItickAppClient(itickCli.Conn()),
	}
}
