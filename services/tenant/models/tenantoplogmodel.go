package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantOpLogModel = (*customTenantOpLogModel)(nil)

type (
	// TenantOpLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantOpLogModel.
	TenantOpLogModel interface {
		tenantOpLogModel
	}

	customTenantOpLogModel struct {
		*defaultTenantOpLogModel
	}
)

// NewTenantOpLogModel returns a model for the database table.
func NewTenantOpLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantOpLogModel {
	return &customTenantOpLogModel{
		defaultTenantOpLogModel: newTenantOpLogModel(conn, c, opts...),
	}
}
