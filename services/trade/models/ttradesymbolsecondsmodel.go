package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSymbolSecondsModel = (*customTTradeSymbolSecondsModel)(nil)

type (
	// TTradeSymbolSecondsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSymbolSecondsModel.
	TTradeSymbolSecondsModel interface {
		tTradeSymbolSecondsModel
	}

	customTTradeSymbolSecondsModel struct {
		*defaultTTradeSymbolSecondsModel
	}
)

// NewTTradeSymbolSecondsModel returns a model for the database table.
func NewTTradeSymbolSecondsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolSecondsModel {
	return &customTTradeSymbolSecondsModel{
		defaultTTradeSymbolSecondsModel: newTTradeSymbolSecondsModel(conn, c, opts...),
	}
}
