package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickProductCategoryModel = (*customTItickProductCategoryModel)(nil)

type (
	// TItickProductCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickProductCategoryModel.
	TItickProductCategoryModel interface {
		tItickProductCategoryModel
	}

	customTItickProductCategoryModel struct {
		*defaultTItickProductCategoryModel
	}
)

// NewTItickProductCategoryModel returns a model for the database table.
func NewTItickProductCategoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickProductCategoryModel {
	return &customTItickProductCategoryModel{
		defaultTItickProductCategoryModel: newTItickProductCategoryModel(conn, c, opts...),
	}
}
