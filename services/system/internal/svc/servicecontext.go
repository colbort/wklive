package svc

import (
	bus "wklive/common/bus/redis"
	"wklive/services/chat/chatinternal"
	"wklive/services/system/internal/config"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/internal/tasks"
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
	TaskPublisher               *bus.Publisher
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
	TenantDomainModel           models.SysTenantDomainModel
	ChatMerchantModel           models.SysChatMerchantModel
	ChatInternal                chatinternal.ChatInternal
}

func NewServiceContext(c config.Config) *ServiceContext {
	taskPublisher := bus.NewPublisherFromRedisConf(c.CacheRedis[0].RedisConf)
	tasks.InitTaskPublisher(taskPublisher)

	conn := sqlx.NewMysql(c.Mysql.DataSource)
	jobLogModel := models.NewSysJobLogModel(conn, c.CacheRedis)
	configModel := models.NewSysConfigModel(conn, c.CacheRedis)
	cron := cronx.NewCronManager(jobLogModel)
	cron.LoadRegisteredHandlers()
	cron.StartScheduler()
	return &ServiceContext{
		Config:                      c,
		DB:                          conn,
		Cache:                       cache.New(c.CacheRedis, syncx.NewSingleFlight(), cache.NewStat(""), redis.Nil),
		Cron:                        cron,
		TaskPublisher:               taskPublisher,
		UserModel:                   models.NewSysUserModel(conn, c.CacheRedis),
		RoleModel:                   models.NewSysRoleModel(conn, c.CacheRedis),
		MenuModel:                   models.NewSysMenuModel(conn, c.CacheRedis),
		UserRoleModel:               models.NewSysUserRoleModel(conn, c.CacheRedis),
		RoleMenuModel:               models.NewSysRoleMenuModel(conn, c.CacheRedis),
		LoginLogModel:               models.NewSysLoginLogModel(conn, c.CacheRedis),
		OpLogModel:                  models.NewSysOpLogModel(conn, c.CacheRedis),
		ConfigModel:                 configModel,
		VerificationCodeRecordModel: models.NewSysVerificationCodeRecordModel(conn, c.CacheRedis),
		JobModel:                    models.NewSysJobModel(conn, c.CacheRedis),
		JobLogModel:                 jobLogModel,
		TenantMode:                  models.NewSysTenantModel(conn, c.CacheRedis),
		TenantDomainModel:           models.NewSysTenantDomainModel(conn),
		ChatMerchantModel:           models.NewSysChatMerchantModel(conn, c.CacheRedis),
		ChatInternal:                chatinternal.NewChatInternal(zrpc.MustNewClient(c.ChatInternalRpc)),
	}
}
