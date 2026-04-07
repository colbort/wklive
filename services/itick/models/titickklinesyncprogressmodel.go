package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickKlineSyncProgressModel = (*customTItickKlineSyncProgressModel)(nil)

type (
	// TItickKlineSyncProgressModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickKlineSyncProgressModel.
	TItickKlineSyncProgressModel interface {
		tItickKlineSyncProgressModel
	}

	customTItickKlineSyncProgressModel struct {
		*defaultTItickKlineSyncProgressModel
	}
)

// NewTItickKlineSyncProgressModel returns a model for the database table.
func NewTItickKlineSyncProgressModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickKlineSyncProgressModel {
	return &customTItickKlineSyncProgressModel{
		defaultTItickKlineSyncProgressModel: newTItickKlineSyncProgressModel(conn, c, opts...),
	}
}
