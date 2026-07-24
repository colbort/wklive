package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityExternalFillModel = (*customTLiquidityExternalFillModel)(nil)

type (
	// TLiquidityExternalFillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityExternalFillModel.
	TLiquidityExternalFillModel interface {
		tLiquidityExternalFillModel
	}

	customTLiquidityExternalFillModel struct {
		*defaultTLiquidityExternalFillModel
	}
)

// NewTLiquidityExternalFillModel returns a model for the database table.
func NewTLiquidityExternalFillModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityExternalFillModel {
	return &customTLiquidityExternalFillModel{
		defaultTLiquidityExternalFillModel: newTLiquidityExternalFillModel(conn, c, opts...),
	}
}
