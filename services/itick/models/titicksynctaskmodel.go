package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickSyncTaskModel = (*customTItickSyncTaskModel)(nil)

type (
	// TItickSyncTaskModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickSyncTaskModel.
	TItickSyncTaskModel interface {
		tItickSyncTaskModel
	}

	customTItickSyncTaskModel struct {
		*defaultTItickSyncTaskModel
	}
)

// NewTItickSyncTaskModel returns a model for the database table.
func NewTItickSyncTaskModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickSyncTaskModel {
	return &customTItickSyncTaskModel{
		defaultTItickSyncTaskModel: newTItickSyncTaskModel(conn, c, opts...),
	}
}
