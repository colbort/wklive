package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantConfigModel = (*customTenantConfigModel)(nil)

type (
	// TenantConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantConfigModel.
	TenantConfigModel interface {
		tenantConfigModel
	}

	customTenantConfigModel struct {
		*defaultTenantConfigModel
	}
)

// NewTenantConfigModel returns a model for the database table.
func NewTenantConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantConfigModel {
	return &customTenantConfigModel{
		defaultTenantConfigModel: newTenantConfigModel(conn, c, opts...),
	}
}
