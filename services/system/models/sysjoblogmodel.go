package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SysJobLogModel = (*customSysJobLogModel)(nil)

type (
	// SysJobLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysJobLogModel.
	SysJobLogModel interface {
		sysJobLogModel
	}

	customSysJobLogModel struct {
		*defaultSysJobLogModel
	}
)

// NewSysJobLogModel returns a model for the database table.
func NewSysJobLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysJobLogModel {
	return &customSysJobLogModel{
		defaultSysJobLogModel: newSysJobLogModel(conn, c, opts...),
	}
}
