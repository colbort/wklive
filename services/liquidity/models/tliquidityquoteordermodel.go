package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityQuoteOrderModel = (*customTLiquidityQuoteOrderModel)(nil)

type (
	// TLiquidityQuoteOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityQuoteOrderModel.
	TLiquidityQuoteOrderModel interface {
		tLiquidityQuoteOrderModel
	}

	customTLiquidityQuoteOrderModel struct {
		*defaultTLiquidityQuoteOrderModel
	}
)

// NewTLiquidityQuoteOrderModel returns a model for the database table.
func NewTLiquidityQuoteOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityQuoteOrderModel {
	return &customTLiquidityQuoteOrderModel{
		defaultTLiquidityQuoteOrderModel: newTLiquidityQuoteOrderModel(conn, c, opts...),
	}
}
