package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TPayOrderModel = (*customTPayOrderModel)(nil)

type (
	// TPayOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPayOrderModel.
	TPayOrderModel interface {
		tPayOrderModel
	}

	customTPayOrderModel struct {
		*defaultTPayOrderModel
	}
)

// NewTPayOrderModel returns a model for the database table.
func NewTPayOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TPayOrderModel {
	return &customTPayOrderModel{
		defaultTPayOrderModel: newTPayOrderModel(conn, c, opts...),
	}
}
