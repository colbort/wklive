package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantUserModel = (*customTenantUserModel)(nil)

type (
	// TenantUserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantUserModel.
	TenantUserModel interface {
		tenantUserModel
	}

	customTenantUserModel struct {
		*defaultTenantUserModel
	}
)

// NewTenantUserModel returns a model for the database table.
func NewTenantUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantUserModel {
	return &customTenantUserModel{
		defaultTenantUserModel: newTenantUserModel(conn, c, opts...),
	}
}
