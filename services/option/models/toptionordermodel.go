package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionOrderModel = (*customTOptionOrderModel)(nil)

type (
	// TOptionOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionOrderModel.
	TOptionOrderModel interface {
		tOptionOrderModel
	}

	customTOptionOrderModel struct {
		*defaultTOptionOrderModel
	}
)

// NewTOptionOrderModel returns a model for the database table.
func NewTOptionOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionOrderModel {
	return &customTOptionOrderModel{
		defaultTOptionOrderModel: newTOptionOrderModel(conn, c, opts...),
	}
}
