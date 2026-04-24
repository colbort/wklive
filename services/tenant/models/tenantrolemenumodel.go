package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantRoleMenuModel = (*customTenantRoleMenuModel)(nil)

type (
	// TenantRoleMenuModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantRoleMenuModel.
	TenantRoleMenuModel interface {
		tenantRoleMenuModel
	}

	customTenantRoleMenuModel struct {
		*defaultTenantRoleMenuModel
	}
)

// NewTenantRoleMenuModel returns a model for the database table.
func NewTenantRoleMenuModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantRoleMenuModel {
	return &customTenantRoleMenuModel{
		defaultTenantRoleMenuModel: newTenantRoleMenuModel(conn, c, opts...),
	}
}
