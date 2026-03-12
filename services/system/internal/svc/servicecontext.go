package svc

import (
	"wklive/services/system/internal/config"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	UserModel     models.UserModel
	RoleModel     models.RoleModel
	MenuModel     models.MenuModel
	UserRoleModel models.UserRoleModel
	RoleMenuModel models.RoleMenuModel
	LoginLogModel models.LoginLogModel
	OpLogModel    models.OpLogModel
	ConfigModel   models.SysConfigModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config: c,

		UserModel:     models.NewSysUserModel(conn).(models.UserModel),
		RoleModel:     models.NewSysRoleModel(conn).(models.RoleModel),
		MenuModel:     models.NewSysMenuModel(conn).(models.MenuModel),
		UserRoleModel: models.NewSysUserRoleModel(conn).(models.UserRoleModel),
		RoleMenuModel: models.NewSysRoleMenuModel(conn).(models.RoleMenuModel),
		LoginLogModel: models.NewSysLoginLogModel(conn).(models.LoginLogModel),
		OpLogModel:    models.NewSysOpLogModel(conn).(models.OpLogModel),
		ConfigModel:   models.NewSysConfigModel(conn),
	}
}
