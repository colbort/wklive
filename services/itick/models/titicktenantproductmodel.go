package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickTenantProductModel = (*customTItickTenantProductModel)(nil)

type (
	// TItickTenantProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickTenantProductModel.
	TItickTenantProductModel interface {
		tItickTenantProductModel
	}

	customTItickTenantProductModel struct {
		*defaultTItickTenantProductModel
	}
)

// NewTItickTenantProductModel returns a model for the database table.
func NewTItickTenantProductModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickTenantProductModel {
	return &customTItickTenantProductModel{
		defaultTItickTenantProductModel: newTItickTenantProductModel(conn, c, opts...),
	}
}
