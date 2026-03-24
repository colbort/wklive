package svc

import (
	"wklive/services/system/internal/config"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	Cron          *cronx.CronManager
	UserModel     models.UserModel
	RoleModel     models.RoleModel
	MenuModel     models.MenuModel
	UserRoleModel models.UserRoleModel
	RoleMenuModel models.RoleMenuModel
	LoginLogModel models.LoginLogModel
	OpLogModel    models.OpLogModel
	ConfigModel   models.ConfigModel
	JobModel      models.JobModel
	JobLogModel   models.JobLogModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	jobLogModel := models.NewSysJobLogModel(conn, c.CacheRedis).(models.JobLogModel)
	cron := cronx.NewCronManager(jobLogModel)
	cron.LoadRegisteredHandlers()
	cron.StartScheduler()
	return &ServiceContext{
		Config:        c,
		Cron:          cron,
		UserModel:     models.NewSysUserModel(conn, c.CacheRedis).(models.UserModel),
		RoleModel:     models.NewSysRoleModel(conn, c.CacheRedis).(models.RoleModel),
		MenuModel:     models.NewSysMenuModel(conn, c.CacheRedis).(models.MenuModel),
		UserRoleModel: models.NewSysUserRoleModel(conn, c.CacheRedis).(models.UserRoleModel),
		RoleMenuModel: models.NewSysRoleMenuModel(conn, c.CacheRedis).(models.RoleMenuModel),
		LoginLogModel: models.NewSysLoginLogModel(conn, c.CacheRedis).(models.LoginLogModel),
		OpLogModel:    models.NewSysOpLogModel(conn, c.CacheRedis).(models.OpLogModel),
		ConfigModel:   models.NewSysConfigModel(conn, c.CacheRedis).(models.ConfigModel),
		JobModel:      models.NewSysJobModel(conn, c.CacheRedis).(models.JobModel),
		JobLogModel:   jobLogModel,
	}
}
