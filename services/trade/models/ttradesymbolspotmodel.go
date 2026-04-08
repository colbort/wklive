package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSymbolSpotModel = (*customTTradeSymbolSpotModel)(nil)

type (
	// TTradeSymbolSpotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSymbolSpotModel.
	TTradeSymbolSpotModel interface {
		tTradeSymbolSpotModel
	}

	customTTradeSymbolSpotModel struct {
		*defaultTTradeSymbolSpotModel
	}
)

// NewTTradeSymbolSpotModel returns a model for the database table.
func NewTTradeSymbolSpotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolSpotModel {
	return &customTTradeSymbolSpotModel{
		defaultTTradeSymbolSpotModel: newTTradeSymbolSpotModel(conn, c, opts...),
	}
}
