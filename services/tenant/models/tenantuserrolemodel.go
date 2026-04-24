package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantUserRoleModel = (*customTenantUserRoleModel)(nil)

type (
	// TenantUserRoleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantUserRoleModel.
	TenantUserRoleModel interface {
		tenantUserRoleModel
	}

	customTenantUserRoleModel struct {
		*defaultTenantUserRoleModel
	}
)

// NewTenantUserRoleModel returns a model for the database table.
func NewTenantUserRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantUserRoleModel {
	return &customTenantUserRoleModel{
		defaultTenantUserRoleModel: newTenantUserRoleModel(conn, c, opts...),
	}
}
