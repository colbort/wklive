package svc

import (
	"wklive/proto/system"
	"wklive/services/user/internal/config"
	"wklive/services/user/models"

	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config            config.Config
	Redis             *redis.Redis
	Node              *snowflake.Node
	UserModel         models.UserModel
	UserSecurityModel models.UserSecurityModel
	UserIdentityModel models.UserIdentityModel
	UserBankModel     models.UserBankModel
	SystemCli         system.SystemClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	node, err := snowflake.NewNode(int64(c.DevServer.Port))
	if err != nil {
		panic(err)
	}
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:            c,
		Redis:             redis.New(c.CacheRedis[0].Host, redis.WithPass(c.CacheRedis[0].Pass)),
		Node:              node,
		UserModel:         models.NewTUserModel(conn, c.CacheRedis).(models.UserModel),
		UserSecurityModel: models.NewTUserSecurityModel(conn, c.CacheRedis).(models.UserSecurityModel),
		UserIdentityModel: models.NewTUserIdentityModel(conn, c.CacheRedis).(models.UserIdentityModel),
		UserBankModel:     models.NewTUserBankModel(conn, c.CacheRedis).(models.UserBankModel),
	}
}
