package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickCategoryModel = (*customTItickCategoryModel)(nil)

type (
	// TItickCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickCategoryModel.
	TItickCategoryModel interface {
		tItickCategoryModel
	}

	customTItickCategoryModel struct {
		*defaultTItickCategoryModel
	}
)

// NewTItickCategoryModel returns a model for the database table.
func NewTItickCategoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickCategoryModel {
	return &customTItickCategoryModel{
		defaultTItickCategoryModel: newTItickCategoryModel(conn, c, opts...),
	}
}
