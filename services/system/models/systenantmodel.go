package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SysTenantModel = (*customSysTenantModel)(nil)

type (
	// SysTenantModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysTenantModel.
	SysTenantModel interface {
		sysTenantModel
	}

	customSysTenantModel struct {
		*defaultSysTenantModel
	}
)

// NewSysTenantModel returns a model for the database table.
func NewSysTenantModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysTenantModel {
	return &customSysTenantModel{
		defaultSysTenantModel: newSysTenantModel(conn, c, opts...),
	}
}
