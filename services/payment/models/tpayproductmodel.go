package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TPayProductModel = (*customTPayProductModel)(nil)

type (
	// TPayProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPayProductModel.
	TPayProductModel interface {
		tPayProductModel
	}

	customTPayProductModel struct {
		*defaultTPayProductModel
	}
)

// NewTPayProductModel returns a model for the database table.
func NewTPayProductModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TPayProductModel {
	return &customTPayProductModel{
		defaultTPayProductModel: newTPayProductModel(conn, c, opts...),
	}
}
