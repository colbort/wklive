package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SysJobModel = (*customSysJobModel)(nil)

type (
	// SysJobModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysJobModel.
	SysJobModel interface {
		sysJobModel
	}

	customSysJobModel struct {
		*defaultSysJobModel
	}
)

// NewSysJobModel returns a model for the database table.
func NewSysJobModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysJobModel {
	return &customSysJobModel{
		defaultSysJobModel: newSysJobModel(conn, c, opts...),
	}
}
