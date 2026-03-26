package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTenantPayPlatformModel = (*customTTenantPayPlatformModel)(nil)

type (
	// TTenantPayPlatformModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTenantPayPlatformModel.
	TTenantPayPlatformModel interface {
		tTenantPayPlatformModel
	}

	customTTenantPayPlatformModel struct {
		*defaultTTenantPayPlatformModel
	}
)

// NewTTenantPayPlatformModel returns a model for the database table.
func NewTTenantPayPlatformModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTenantPayPlatformModel {
	return &customTTenantPayPlatformModel{
		defaultTTenantPayPlatformModel: newTTenantPayPlatformModel(conn, c, opts...),
	}
}
