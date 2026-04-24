package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantLoginLogModel = (*customTenantLoginLogModel)(nil)

type (
	// TenantLoginLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantLoginLogModel.
	TenantLoginLogModel interface {
		tenantLoginLogModel
	}

	customTenantLoginLogModel struct {
		*defaultTenantLoginLogModel
	}
)

// NewTenantLoginLogModel returns a model for the database table.
func NewTenantLoginLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantLoginLogModel {
	return &customTenantLoginLogModel{
		defaultTenantLoginLogModel: newTenantLoginLogModel(conn, c, opts...),
	}
}
