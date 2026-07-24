package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityReconcileDetailModel = (*customTLiquidityReconcileDetailModel)(nil)

type (
	// TLiquidityReconcileDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityReconcileDetailModel.
	TLiquidityReconcileDetailModel interface {
		tLiquidityReconcileDetailModel
	}

	customTLiquidityReconcileDetailModel struct {
		*defaultTLiquidityReconcileDetailModel
	}
)

// NewTLiquidityReconcileDetailModel returns a model for the database table.
func NewTLiquidityReconcileDetailModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityReconcileDetailModel {
	return &customTLiquidityReconcileDetailModel{
		defaultTLiquidityReconcileDetailModel: newTLiquidityReconcileDetailModel(conn, c, opts...),
	}
}
