// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"wklive/payment-api/internal/config"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	PaymentCli payment.CallbackClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := zrpc.MustNewClient(c.PaymentRpc)
	return &ServiceContext{
		Config:     c,
		PaymentCli: payment.NewCallbackClient(client.Conn()),
	}
}
