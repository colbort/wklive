package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickTenantProductCategoryModel = (*customTItickTenantProductCategoryModel)(nil)

type (
	// TItickTenantProductCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickTenantProductCategoryModel.
	TItickTenantProductCategoryModel interface {
		tItickTenantProductCategoryModel
	}

	customTItickTenantProductCategoryModel struct {
		*defaultTItickTenantProductCategoryModel
	}
)

// NewTItickTenantProductCategoryModel returns a model for the database table.
func NewTItickTenantProductCategoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickTenantProductCategoryModel {
	return &customTItickTenantProductCategoryModel{
		defaultTItickTenantProductCategoryModel: newTItickTenantProductCategoryModel(conn, c, opts...),
	}
}
