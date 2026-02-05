package models

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SysRoleMenuModel = (*customSysRoleMenuModel)(nil)

type (
	// SysRoleMenuModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysRoleMenuModel.
	SysRoleMenuModel interface {
		sysRoleMenuModel
		withSession(session sqlx.Session) SysRoleMenuModel
	}

	customSysRoleMenuModel struct {
		*defaultSysRoleMenuModel
	}
)

// NewSysRoleMenuModel returns a model for the database table.
func NewSysRoleMenuModel(conn sqlx.SqlConn) SysRoleMenuModel {
	return &customSysRoleMenuModel{
		defaultSysRoleMenuModel: newSysRoleMenuModel(conn),
	}
}

func (m *customSysRoleMenuModel) withSession(session sqlx.Session) SysRoleMenuModel {
	return NewSysRoleMenuModel(sqlx.NewSqlConnFromSession(session))
}
