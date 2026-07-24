package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityInventorySnapshotModel = (*customTLiquidityInventorySnapshotModel)(nil)

type (
	// TLiquidityInventorySnapshotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityInventorySnapshotModel.
	TLiquidityInventorySnapshotModel interface {
		tLiquidityInventorySnapshotModel
	}

	customTLiquidityInventorySnapshotModel struct {
		*defaultTLiquidityInventorySnapshotModel
	}
)

// NewTLiquidityInventorySnapshotModel returns a model for the database table.
func NewTLiquidityInventorySnapshotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityInventorySnapshotModel {
	return &customTLiquidityInventorySnapshotModel{
		defaultTLiquidityInventorySnapshotModel: newTLiquidityInventorySnapshotModel(conn, c, opts...),
	}
}
