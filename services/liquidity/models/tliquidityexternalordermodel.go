package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityExternalOrderModel = (*customTLiquidityExternalOrderModel)(nil)

type (
	// TLiquidityExternalOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityExternalOrderModel.
	TLiquidityExternalOrderModel interface {
		tLiquidityExternalOrderModel
	}

	customTLiquidityExternalOrderModel struct {
		*defaultTLiquidityExternalOrderModel
	}
)

// NewTLiquidityExternalOrderModel returns a model for the database table.
func NewTLiquidityExternalOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityExternalOrderModel {
	return &customTLiquidityExternalOrderModel{
		defaultTLiquidityExternalOrderModel: newTLiquidityExternalOrderModel(conn, c, opts...),
	}
}
