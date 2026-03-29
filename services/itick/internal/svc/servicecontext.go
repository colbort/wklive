package svc

import (
	"wklive/proto/system"
	"wklive/services/itick/internal/config"
	"wklive/services/itick/internal/socket/client"
	"wklive/services/itick/internal/socket/server"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	SystemCli          system.SystemClient
	ItickManager       *client.ItickManager
	ItickCategoryModel models.ItickCategoryModel
	Hub                *server.Hub
}

func NewServiceContext(c config.Config) *ServiceContext {
	hub := server.NewHub()

	conn := sqlx.NewMysql(c.Mysql.DataSource)
	itickCategoryModel := models.NewTItickCategoryModel(conn, c.CacheRedis).(models.ItickCategoryModel)
	itickManager := client.NewItickManager(c.Itick.WSUrl, c.Itick.Token, hub, itickCategoryModel)
	hub.SetHooks(
		func(key string, msg server.ClientMessage) {
			if err := itickManager.Subscribe(msg); err != nil {
				logx.Errorf("subscribe upstream failed, key=%s, err=%v", key, err)
			}
		},
		func(key string, msg server.ClientMessage) {
			if err := itickManager.Unsubscribe(msg); err != nil {
				logx.Errorf("unsubscribe upstream failed, key=%s, err=%v", key, err)
			}
		},
	)

	return &ServiceContext{
		Config:             c,
		ItickManager:       itickManager,
		ItickCategoryModel: itickCategoryModel,
		Hub:                hub,
	}
}
