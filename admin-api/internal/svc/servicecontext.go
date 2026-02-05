// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"wklive/admin-api/internal/config"
	"wklive/rpc/system"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	SystemCli system.SystemClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	cli := zrpc.MustNewClient(c.SystemRpc)
	return &ServiceContext{
		Config:    c,
		SystemCli: system.NewSystemClient(cli.Conn()),
	}
}
