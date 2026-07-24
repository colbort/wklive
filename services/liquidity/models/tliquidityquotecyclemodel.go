package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityQuoteCycleModel = (*customTLiquidityQuoteCycleModel)(nil)

type (
	// TLiquidityQuoteCycleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityQuoteCycleModel.
	TLiquidityQuoteCycleModel interface {
		tLiquidityQuoteCycleModel
	}

	customTLiquidityQuoteCycleModel struct {
		*defaultTLiquidityQuoteCycleModel
	}
)

// NewTLiquidityQuoteCycleModel returns a model for the database table.
func NewTLiquidityQuoteCycleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityQuoteCycleModel {
	return &customTLiquidityQuoteCycleModel{
		defaultTLiquidityQuoteCycleModel: newTLiquidityQuoteCycleModel(conn, c, opts...),
	}
}
