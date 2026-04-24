package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantMenuModel = (*customTenantMenuModel)(nil)

type (
	// TenantMenuModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantMenuModel.
	TenantMenuModel interface {
		tenantMenuModel
	}

	customTenantMenuModel struct {
		*defaultTenantMenuModel
	}
)

// NewTenantMenuModel returns a model for the database table.
func NewTenantMenuModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantMenuModel {
	return &customTenantMenuModel{
		defaultTenantMenuModel: newTenantMenuModel(conn, c, opts...),
	}
}
