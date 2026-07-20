package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeOrderSecondsModel = (*customTTradeOrderSecondsModel)(nil)

type (
	// TTradeOrderSecondsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeOrderSecondsModel.
	TTradeOrderSecondsModel interface {
		tTradeOrderSecondsModel
	}

	customTTradeOrderSecondsModel struct {
		*defaultTTradeOrderSecondsModel
	}
)

// NewTTradeOrderSecondsModel returns a model for the database table.
func NewTTradeOrderSecondsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeOrderSecondsModel {
	return &customTTradeOrderSecondsModel{
		defaultTTradeOrderSecondsModel: newTTradeOrderSecondsModel(conn, c, opts...),
	}
}
