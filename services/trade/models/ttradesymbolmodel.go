package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSymbolModel = (*customTTradeSymbolModel)(nil)

type (
	// TTradeSymbolModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSymbolModel.
	TTradeSymbolModel interface {
		tTradeSymbolModel
	}

	customTTradeSymbolModel struct {
		*defaultTTradeSymbolModel
	}
)

// NewTTradeSymbolModel returns a model for the database table.
func NewTTradeSymbolModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolModel {
	return &customTTradeSymbolModel{
		defaultTTradeSymbolModel: newTTradeSymbolModel(conn, c, opts...),
	}
}
