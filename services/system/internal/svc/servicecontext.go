package svc

import (
	"wklive/proto/itick"
	"wklive/proto/option"
	"wklive/proto/staking"
	"wklive/proto/trade"
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
	UserModel                   models.UserModel
	RoleModel                   models.RoleModel
	MenuModel                   models.MenuModel
	UserRoleModel               models.UserRoleModel
	RoleMenuModel               models.RoleMenuModel
	LoginLogModel               models.LoginLogModel
	OpLogModel                  models.OpLogModel
	ConfigModel                 models.ConfigModel
	VerificationCodeRecordModel models.VerificationCodeRecordModel
	JobModel                    models.JobModel
	JobLogModel                 models.JobLogModel
	TenantMode                  models.TenantModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	global.ItickTaskCli = itick.NewItickTaskClient(zrpc.MustNewClient(c.ItickRpc).Conn())
	global.OptionTaskCli = option.NewOptionTaskClient(zrpc.MustNewClient(c.OptionRpc).Conn())
	global.StakingTaskCli = staking.NewStakingTaskClient(zrpc.MustNewClient(c.StakingRpc).Conn())
	global.TradeTaskCli = trade.NewTradeTaskClient(zrpc.MustNewClient(c.TradeRpc).Conn())

	conn := sqlx.NewMysql(c.Mysql.DataSource)
	jobLogModel := models.NewSysJobLogModel(conn, c.CacheRedis).(models.JobLogModel)
	global.ConfigModel = models.NewSysConfigModel(conn, c.CacheRedis).(models.ConfigModel)
	cron := cronx.NewCronManager(jobLogModel)
	cron.LoadRegisteredHandlers()
	cron.StartScheduler()
	return &ServiceContext{
		Config:                      c,
		DB:                          conn,
		Cache:                       cache.New(c.CacheRedis, syncx.NewSingleFlight(), cache.NewStat(""), redis.Nil),
		Cron:                        cron,
		UserModel:                   models.NewSysUserModel(conn, c.CacheRedis).(models.UserModel),
		RoleModel:                   models.NewSysRoleModel(conn, c.CacheRedis).(models.RoleModel),
		MenuModel:                   models.NewSysMenuModel(conn, c.CacheRedis).(models.MenuModel),
		UserRoleModel:               models.NewSysUserRoleModel(conn, c.CacheRedis).(models.UserRoleModel),
		RoleMenuModel:               models.NewSysRoleMenuModel(conn, c.CacheRedis).(models.RoleMenuModel),
		LoginLogModel:               models.NewSysLoginLogModel(conn, c.CacheRedis).(models.LoginLogModel),
		OpLogModel:                  models.NewSysOpLogModel(conn, c.CacheRedis).(models.OpLogModel),
		ConfigModel:                 global.ConfigModel,
		VerificationCodeRecordModel: models.NewSysVerificationCodeRecordModel(conn, c.CacheRedis).(models.VerificationCodeRecordModel),
		JobModel:                    models.NewSysJobModel(conn, c.CacheRedis).(models.JobModel),
		JobLogModel:                 jobLogModel,
		TenantMode:                  models.NewSysTenantModel(conn, c.CacheRedis).(models.TenantModel),
	}
}
