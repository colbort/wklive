package svc

import (
	"wklive/proto/system"
	"wklive/services/user/internal/config"
	"wklive/services/user/models"

	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config            config.Config
	DB                sqlx.SqlConn
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
	systemCli := zrpc.MustNewClient(c.SystemRpc)
	return &ServiceContext{
		Config:            c,
		DB:                conn,
		Redis:             redis.MustNewRedis(c.Redis.RedisConf),
		Node:              node,
		UserModel:         models.NewTUserModel(conn, c.CacheRedis).(models.UserModel),
		UserSecurityModel: models.NewTUserSecurityModel(conn, c.CacheRedis).(models.UserSecurityModel),
		UserIdentityModel: models.NewTUserIdentityModel(conn, c.CacheRedis).(models.UserIdentityModel),
		UserBankModel:     models.NewTUserBankModel(conn, c.CacheRedis).(models.UserBankModel),
		SystemCli:         system.NewSystemClient(systemCli.Conn()),
	}
}
