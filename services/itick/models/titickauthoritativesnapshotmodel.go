package models

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickAuthoritativeSnapshotModel = (*customTItickAuthoritativeSnapshotModel)(nil)

type (
	// TItickAuthoritativeSnapshotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickAuthoritativeSnapshotModel.
	TItickAuthoritativeSnapshotModel interface {
		tItickAuthoritativeSnapshotModel
		InsertIgnore(context.Context, *TItickAuthoritativeSnapshot) error
		FindAtOrBefore(context.Context, string, string, string, string, int64, int64) (*TItickAuthoritativeSnapshot, error)
	}

	customTItickAuthoritativeSnapshotModel struct {
		*defaultTItickAuthoritativeSnapshotModel
	}
)

// NewTItickAuthoritativeSnapshotModel returns a model for the database table.
func NewTItickAuthoritativeSnapshotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickAuthoritativeSnapshotModel {
	return &customTItickAuthoritativeSnapshotModel{
		defaultTItickAuthoritativeSnapshotModel: newTItickAuthoritativeSnapshotModel(conn, c, opts...),
	}
}

func (m *defaultTItickAuthoritativeSnapshotModel) InsertIgnore(ctx context.Context, row *TItickAuthoritativeSnapshot) error {
	_, err := m.ExecNoCacheCtx(ctx, `INSERT IGNORE INTO t_itick_authoritative_snapshot
(snapshot_id,authority,snapshot_kind,category_code,market,symbol,price,source_timestamp,snapshot_timestamp,revision,formula_version,raw_payload,create_times)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, row.SnapshotId, row.Authority, row.SnapshotKind, row.CategoryCode, row.Market, row.Symbol, row.Price, row.SourceTimestamp, row.SnapshotTimestamp, row.Revision, row.FormulaVersion, row.RawPayload, row.CreateTimes)
	return err
}

func (m *defaultTItickAuthoritativeSnapshotModel) FindAtOrBefore(ctx context.Context, authority, category, market, symbol string, targetTime, minTime int64) (*TItickAuthoritativeSnapshot, error) {
	var row TItickAuthoritativeSnapshot
	err := m.QueryRowNoCacheCtx(ctx, &row, `SELECT id,snapshot_id,authority,snapshot_kind,category_code,market,symbol,price,source_timestamp,snapshot_timestamp,revision,formula_version,raw_payload,create_times
FROM t_itick_authoritative_snapshot
WHERE authority=? AND category_code=? AND market=? AND symbol=? AND source_timestamp<=? AND source_timestamp>=?
ORDER BY source_timestamp DESC,revision DESC LIMIT 1`, authority, category, market, symbol, targetTime, minTime)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
