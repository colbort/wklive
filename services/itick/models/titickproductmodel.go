package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickProductModel = (*customTItickProductModel)(nil)

type (
	// TItickProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickProductModel.
	TItickProductModel interface {
		tItickProductModel
	}

	customTItickProductModel struct {
		*defaultTItickProductModel
	}
)

// NewTItickProductModel returns a model for the database table.
func NewTItickProductModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickProductModel {
	return &customTItickProductModel{
		defaultTItickProductModel: newTItickProductModel(conn, c, opts...),
	}
}
