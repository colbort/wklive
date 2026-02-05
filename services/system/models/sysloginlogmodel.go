package models

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SysLoginLogModel = (*customSysLoginLogModel)(nil)

type (
	// SysLoginLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysLoginLogModel.
	SysLoginLogModel interface {
		sysLoginLogModel
		withSession(session sqlx.Session) SysLoginLogModel
	}

	customSysLoginLogModel struct {
		*defaultSysLoginLogModel
	}
)

// NewSysLoginLogModel returns a model for the database table.
func NewSysLoginLogModel(conn sqlx.SqlConn) SysLoginLogModel {
	return &customSysLoginLogModel{
		defaultSysLoginLogModel: newSysLoginLogModel(conn),
	}
}

func (m *customSysLoginLogModel) withSession(session sqlx.Session) SysLoginLogModel {
	return NewSysLoginLogModel(sqlx.NewSqlConnFromSession(session))
}
