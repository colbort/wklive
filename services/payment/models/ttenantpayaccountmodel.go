package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTenantPayAccountModel = (*customTTenantPayAccountModel)(nil)

type (
	// TTenantPayAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTenantPayAccountModel.
	TTenantPayAccountModel interface {
		tTenantPayAccountModel
	}

	customTTenantPayAccountModel struct {
		*defaultTTenantPayAccountModel
	}
)

// NewTTenantPayAccountModel returns a model for the database table.
func NewTTenantPayAccountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTenantPayAccountModel {
	return &customTTenantPayAccountModel{
		defaultTTenantPayAccountModel: newTTenantPayAccountModel(conn, c, opts...),
	}
}
