// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"wklive/app-api/internal/config"

	"wklive/proto/payment"
	"wklive/proto/user"
)

type ServiceContext struct {
	Config     config.Config
	UserCli    user.UserAppClient
	PaymentCli payment.PaymentAppClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
