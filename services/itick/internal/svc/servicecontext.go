package svc

import (
	"wklive/proto/system"
	"wklive/services/itick/internal/config"
	"wklive/services/itick/internal/socket/client"
	"wklive/services/itick/internal/socket/server"
)

type ServiceContext struct {
	Config       config.Config
	SystemCli    system.SystemClient
	ItickManager *client.ItickManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	hub := server.NewHub()
	hub.SetHooks(
		func(key string, msg server.ClientMessage) {
		},
		func(key string, msg server.ClientMessage) {
		},
	)
	itickManager := client.NewItickManager(c.Itick.Token, hub)

	return &ServiceContext{
		Config:       c,
		ItickManager: itickManager,
	}
}
