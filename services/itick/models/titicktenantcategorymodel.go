package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickTenantCategoryModel = (*customTItickTenantCategoryModel)(nil)

type (
	// TItickTenantCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickTenantCategoryModel.
	TItickTenantCategoryModel interface {
		tItickTenantCategoryModel
	}

	customTItickTenantCategoryModel struct {
		*defaultTItickTenantCategoryModel
	}
)

// NewTItickTenantCategoryModel returns a model for the database table.
func NewTItickTenantCategoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickTenantCategoryModel {
	return &customTItickTenantCategoryModel{
		defaultTItickTenantCategoryModel: newTItickTenantCategoryModel(conn, c, opts...),
	}
}
