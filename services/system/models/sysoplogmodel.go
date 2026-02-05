package models

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SysOpLogModel = (*customSysOpLogModel)(nil)

type (
	// SysOpLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysOpLogModel.
	SysOpLogModel interface {
		sysOpLogModel
		withSession(session sqlx.Session) SysOpLogModel
	}

	customSysOpLogModel struct {
		*defaultSysOpLogModel
	}
)

// NewSysOpLogModel returns a model for the database table.
func NewSysOpLogModel(conn sqlx.SqlConn) SysOpLogModel {
	return &customSysOpLogModel{
		defaultSysOpLogModel: newSysOpLogModel(conn),
	}
}

func (m *customSysOpLogModel) withSession(session sqlx.Session) SysOpLogModel {
	return NewSysOpLogModel(sqlx.NewSqlConnFromSession(session))
}
