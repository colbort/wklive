package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TPayPlatformModel = (*customTPayPlatformModel)(nil)

type (
	// TPayPlatformModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPayPlatformModel.
	TPayPlatformModel interface {
		tPayPlatformModel
	}

	customTPayPlatformModel struct {
		*defaultTPayPlatformModel
	}
)

// NewTPayPlatformModel returns a model for the database table.
func NewTPayPlatformModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TPayPlatformModel {
	return &customTPayPlatformModel{
		defaultTPayPlatformModel: newTPayPlatformModel(conn, c, opts...),
	}
}
