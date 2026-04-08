package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeOrderModel = (*customTTradeOrderModel)(nil)

type (
	// TTradeOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeOrderModel.
	TTradeOrderModel interface {
		tTradeOrderModel
	}

	customTTradeOrderModel struct {
		*defaultTTradeOrderModel
	}
)

// NewTTradeOrderModel returns a model for the database table.
func NewTTradeOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeOrderModel {
	return &customTTradeOrderModel{
		defaultTTradeOrderModel: newTTradeOrderModel(conn, c, opts...),
	}
}
