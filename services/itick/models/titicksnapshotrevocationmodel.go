package models

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickSnapshotRevocationModel = (*customTItickSnapshotRevocationModel)(nil)

type (
	// TItickSnapshotRevocationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickSnapshotRevocationModel.
	TItickSnapshotRevocationModel interface {
		tItickSnapshotRevocationModel
		FindAfterID(context.Context, int64, int64) ([]*TItickSnapshotRevocation, error)
	}

	customTItickSnapshotRevocationModel struct {
		*defaultTItickSnapshotRevocationModel
	}
)

// NewTItickSnapshotRevocationModel returns a model for the database table.
func NewTItickSnapshotRevocationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickSnapshotRevocationModel {
	return &customTItickSnapshotRevocationModel{
		defaultTItickSnapshotRevocationModel: newTItickSnapshotRevocationModel(conn, c, opts...),
	}
}

func (m *defaultTItickSnapshotRevocationModel) FindAfterID(ctx context.Context, afterID, limit int64) ([]*TItickSnapshotRevocation, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	var rows []*TItickSnapshotRevocation
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tItickSnapshotRevocationRows+" FROM t_itick_snapshot_revocation WHERE id>? ORDER BY id LIMIT ?", afterID, limit)
	return rows, err
}
