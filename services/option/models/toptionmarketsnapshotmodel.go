package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionMarketSnapshotModel = (*customTOptionMarketSnapshotModel)(nil)

type (
	// TOptionMarketSnapshotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionMarketSnapshotModel.
	TOptionMarketSnapshotModel interface {
		tOptionMarketSnapshotModel
	}

	customTOptionMarketSnapshotModel struct {
		*defaultTOptionMarketSnapshotModel
	}
)

// NewTOptionMarketSnapshotModel returns a model for the database table.
func NewTOptionMarketSnapshotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionMarketSnapshotModel {
	return &customTOptionMarketSnapshotModel{
		defaultTOptionMarketSnapshotModel: newTOptionMarketSnapshotModel(conn, c, opts...),
	}
}
