package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityHedgeTaskModel = (*customTLiquidityHedgeTaskModel)(nil)

type (
	// TLiquidityHedgeTaskModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityHedgeTaskModel.
	TLiquidityHedgeTaskModel interface {
		tLiquidityHedgeTaskModel
	}

	customTLiquidityHedgeTaskModel struct {
		*defaultTLiquidityHedgeTaskModel
	}
)

// NewTLiquidityHedgeTaskModel returns a model for the database table.
func NewTLiquidityHedgeTaskModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityHedgeTaskModel {
	return &customTLiquidityHedgeTaskModel{
		defaultTLiquidityHedgeTaskModel: newTLiquidityHedgeTaskModel(conn, c, opts...),
	}
}
