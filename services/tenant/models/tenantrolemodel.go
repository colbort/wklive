package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantRoleModel = (*customTenantRoleModel)(nil)

type (
	// TenantRoleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantRoleModel.
	TenantRoleModel interface {
		tenantRoleModel
	}

	customTenantRoleModel struct {
		*defaultTenantRoleModel
	}
)

// NewTenantRoleModel returns a model for the database table.
func NewTenantRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantRoleModel {
	return &customTenantRoleModel{
		defaultTenantRoleModel: newTenantRoleModel(conn, c, opts...),
	}
}
