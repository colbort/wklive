package svc

import (
	"wklive/proto/itick"
	"wklive/proto/option"
	"wklive/proto/staking"
	"wklive/proto/trade"
	"wklive/services/chat/chatinternal"
	"wklive/services/system/internal/config"
	"wklive/services/system/internal/global"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                      config.Config
	DB                          sqlx.SqlConn
	Cache                       cache.Cache
	Cron                        *cronx.CronManager
	UserModel                   models.SysUserModel
	RoleModel                   models.SysRoleModel
	MenuModel                   models.SysMenuModel
	UserRoleModel               models.SysUserRoleModel
	RoleMenuModel               models.SysRoleMenuModel
	LoginLogModel               models.SysLoginLogModel
	OpLogModel                  models.SysOpLogModel
	ConfigModel                 models.SysConfigModel
	VerificationCodeRecordModel models.SysVerificationCodeRecordModel
	JobModel                    models.SysJobModel
	JobLogModel                 models.SysJobLogModel
	TenantMode                  models.SysTenantModel
	ChatMerchantModel           models.SysChatMerchantModel
	ChatInternal                chatinternal.ChatInternal
}

func NewServiceContext(c config.Config) *ServiceContext {
	global.ItickTaskCli = itick.NewItickTaskClient(zrpc.MustNewClient(c.ItickRpc).Conn())
	global.OptionTaskCli = option.NewOptionTaskClient(zrpc.MustNewClient(c.OptionRpc).Conn())
	global.StakingTaskCli = staking.NewStakingTaskClient(zrpc.MustNewClient(c.StakingRpc).Conn())
	global.TradeTaskCli = trade.NewTradeTaskClient(zrpc.MustNewClient(c.TradeRpc).Conn())

	conn := sqlx.NewMysql(c.Mysql.DataSource)
	jobLogModel := models.NewSysJobLogModel(conn, c.CacheRedis)
	global.ConfigModel = models.NewSysConfigModel(conn, c.CacheRedis)
	cron := cronx.NewCronManager(jobLogModel)
	cron.LoadRegisteredHandlers()
	cron.StartScheduler()
	return &ServiceContext{
		Config:                      c,
		DB:                          conn,
		Cache:                       cache.New(c.CacheRedis, syncx.NewSingleFlight(), cache.NewStat(""), redis.Nil),
		Cron:                        cron,
		UserModel:                   models.NewSysUserModel(conn, c.CacheRedis),
		RoleModel:                   models.NewSysRoleModel(conn, c.CacheRedis),
		MenuModel:                   models.NewSysMenuModel(conn, c.CacheRedis),
		UserRoleModel:               models.NewSysUserRoleModel(conn, c.CacheRedis),
		RoleMenuModel:               models.NewSysRoleMenuModel(conn, c.CacheRedis),
		LoginLogModel:               models.NewSysLoginLogModel(conn, c.CacheRedis),
		OpLogModel:                  models.NewSysOpLogModel(conn, c.CacheRedis),
		ConfigModel:                 global.ConfigModel,
		VerificationCodeRecordModel: models.NewSysVerificationCodeRecordModel(conn, c.CacheRedis),
		JobModel:                    models.NewSysJobModel(conn, c.CacheRedis),
		JobLogModel:                 jobLogModel,
		TenantMode:                  models.NewSysTenantModel(conn, c.CacheRedis),
		ChatMerchantModel:           models.NewSysChatMerchantModel(conn, c.CacheRedis),
		ChatInternal:                chatinternal.NewChatInternal(zrpc.MustNewClient(c.ChatInternalRpc)),
	}
}
