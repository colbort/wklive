package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTenantModel = (*customTTenantModel)(nil)

type (
	// TTenantModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTenantModel.
	TTenantModel interface {
		tTenantModel
	}

	customTTenantModel struct {
		*defaultTTenantModel
	}
)

// NewTTenantModel returns a model for the database table.
func NewTTenantModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTenantModel {
	return &customTTenantModel{
		defaultTTenantModel: newTTenantModel(conn, c, opts...),
	}
}
