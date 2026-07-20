package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSymbolLeverageDefaultModel = (*customTTradeSymbolLeverageDefaultModel)(nil)

type (
	// TTradeSymbolLeverageDefaultModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSymbolLeverageDefaultModel.
	TTradeSymbolLeverageDefaultModel interface {
		tTradeSymbolLeverageDefaultModel
	}

	customTTradeSymbolLeverageDefaultModel struct {
		*defaultTTradeSymbolLeverageDefaultModel
	}
)

// NewTTradeSymbolLeverageDefaultModel returns a model for the database table.
func NewTTradeSymbolLeverageDefaultModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolLeverageDefaultModel {
	return &customTTradeSymbolLeverageDefaultModel{
		defaultTTradeSymbolLeverageDefaultModel: newTTradeSymbolLeverageDefaultModel(conn, c, opts...),
	}
}
