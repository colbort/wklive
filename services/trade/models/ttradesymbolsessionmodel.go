package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSymbolSessionModel = (*customTTradeSymbolSessionModel)(nil)

type (
	// TTradeSymbolSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSymbolSessionModel.
	TTradeSymbolSessionModel interface {
		tTradeSymbolSessionModel
	}

	customTTradeSymbolSessionModel struct {
		*defaultTTradeSymbolSessionModel
	}
)

// NewTTradeSymbolSessionModel returns a model for the database table.
func NewTTradeSymbolSessionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolSessionModel {
	return &customTTradeSymbolSessionModel{
		defaultTTradeSymbolSessionModel: newTTradeSymbolSessionModel(conn, c, opts...),
	}
}
