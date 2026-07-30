package models

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TMarketSnapshotRevocationModel = (*customTMarketSnapshotRevocationModel)(nil)

type (
	// TMarketSnapshotRevocationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTMarketSnapshotRevocationModel.
	TMarketSnapshotRevocationModel interface {
		tMarketSnapshotRevocationModel
		FindAfterID(context.Context, int64, int64) ([]*TMarketSnapshotRevocation, error)
	}

	customTMarketSnapshotRevocationModel struct {
		*defaultTMarketSnapshotRevocationModel
	}
)

// NewTMarketSnapshotRevocationModel returns a model for the database table.
func NewTMarketSnapshotRevocationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TMarketSnapshotRevocationModel {
	return &customTMarketSnapshotRevocationModel{
		defaultTMarketSnapshotRevocationModel: newTMarketSnapshotRevocationModel(conn, c, opts...),
	}
}

func (m *defaultTMarketSnapshotRevocationModel) FindAfterID(ctx context.Context, afterID, limit int64) ([]*TMarketSnapshotRevocation, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	var rows []*TMarketSnapshotRevocation
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tMarketSnapshotRevocationRows+" FROM t_itick_snapshot_revocation WHERE id>? ORDER BY id LIMIT ?", afterID, limit)
	return rows, err
}
