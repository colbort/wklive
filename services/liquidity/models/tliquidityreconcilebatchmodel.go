package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityReconcileBatchModel = (*customTLiquidityReconcileBatchModel)(nil)

type (
	// TLiquidityReconcileBatchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityReconcileBatchModel.
	TLiquidityReconcileBatchModel interface {
		tLiquidityReconcileBatchModel
	}

	customTLiquidityReconcileBatchModel struct {
		*defaultTLiquidityReconcileBatchModel
	}
)

// NewTLiquidityReconcileBatchModel returns a model for the database table.
func NewTLiquidityReconcileBatchModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityReconcileBatchModel {
	return &customTLiquidityReconcileBatchModel{
		defaultTLiquidityReconcileBatchModel: newTLiquidityReconcileBatchModel(conn, c, opts...),
	}
}
