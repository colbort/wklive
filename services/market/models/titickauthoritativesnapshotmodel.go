package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickAuthoritativeSnapshotModel = (*customTItickAuthoritativeSnapshotModel)(nil)

type (
	// TItickAuthoritativeSnapshotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickAuthoritativeSnapshotModel.
	TItickAuthoritativeSnapshotModel interface {
		tItickAuthoritativeSnapshotModel
		InsertImmutable(context.Context, *TItickAuthoritativeSnapshot) error
		InsertImmutableAndEnqueue(context.Context, *TItickAuthoritativeSnapshot, string) error
		FindAtOrBefore(context.Context, string, string, string, string, string, int64, int64) (*TItickAuthoritativeSnapshot, error)
		FindAfterID(context.Context, int64, int64) ([]*TItickAuthoritativeSnapshot, error)
		FindProductKeys(context.Context) ([]AuthoritativeSnapshotProductKey, error)
	}

	AuthoritativeSnapshotProductKey struct {
		Authority    string `db:"authority"`
		SnapshotKind string `db:"snapshot_kind"`
		CategoryCode string `db:"category_code"`
		Market       string `db:"market"`
		Symbol       string `db:"symbol"`
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

func (m *customTItickAuthoritativeSnapshotModel) FindProductKeys(ctx context.Context) ([]AuthoritativeSnapshotProductKey, error) {
	var rows []AuthoritativeSnapshotProductKey
	err := m.QueryRowsNoCacheCtx(ctx, &rows, `SELECT DISTINCT authority,snapshot_kind,category_code,market,symbol
FROM t_itick_authoritative_snapshot
ORDER BY authority,snapshot_kind,category_code,market,symbol`)
	return rows, err
}

func (m *customTItickAuthoritativeSnapshotModel) FindAfterID(ctx context.Context, afterID, limit int64) ([]*TItickAuthoritativeSnapshot, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	var rows []*TItickAuthoritativeSnapshot
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tItickAuthoritativeSnapshotRows+" FROM t_itick_authoritative_snapshot WHERE id>? ORDER BY id LIMIT ?", afterID, limit)
	return rows, err
}

func (m *customTItickAuthoritativeSnapshotModel) InsertImmutableAndEnqueue(ctx context.Context, row *TItickAuthoritativeSnapshot, payload string) error {
	return m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		result, err := conn.ExecCtx(ctx, `INSERT INTO t_itick_authoritative_snapshot
(snapshot_id,authority,snapshot_kind,category_code,market,symbol,price,source_timestamp,snapshot_timestamp,revision,formula_version,raw_payload,create_times)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE snapshot_id=snapshot_id`, row.SnapshotId, row.Authority, row.SnapshotKind, row.CategoryCode, row.Market, row.Symbol, row.Price, row.SourceTimestamp, row.SnapshotTimestamp, row.Revision, row.FormulaVersion, row.RawPayload, row.CreateTimes)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			var existing TItickAuthoritativeSnapshot
			query := `SELECT id,snapshot_id,authority,snapshot_kind,category_code,market,symbol,price,source_timestamp,snapshot_timestamp,revision,formula_version,raw_payload,create_times
FROM t_itick_authoritative_snapshot WHERE snapshot_id=? LIMIT 1`
			if findErr := conn.QueryRowCtx(ctx, &existing, query, row.SnapshotId); findErr != nil {
				return findErr
			}
			if sameAuthoritativeSnapshotIdentity(&existing, row) {
				// The permanent archive is the deduplication source after a successful
				// outbox row has been cleaned up. Do not enqueue a replayed snapshot.
				return nil
			}
			return authoritativeSnapshotConflictError(row)
		}
		_, err = conn.ExecCtx(ctx, "INSERT INTO t_itick_snapshot_outbox(snapshot_id,payload,status,retry_count,next_retry_at,last_error_msg,create_times,update_times) VALUES(?,?,1,0,0,'',?,?)", row.SnapshotId, payload, row.CreateTimes, row.CreateTimes)
		return err
	})
}

func (m *customTItickAuthoritativeSnapshotModel) InsertImmutable(ctx context.Context, row *TItickAuthoritativeSnapshot) error {
	result, err := m.ExecNoCacheCtx(ctx, `INSERT INTO t_itick_authoritative_snapshot
(snapshot_id,authority,snapshot_kind,category_code,market,symbol,price,source_timestamp,snapshot_timestamp,revision,formula_version,raw_payload,create_times)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE snapshot_id=snapshot_id`, row.SnapshotId, row.Authority, row.SnapshotKind, row.CategoryCode, row.Market, row.Symbol, row.Price, row.SourceTimestamp, row.SnapshotTimestamp, row.Revision, row.FormulaVersion, row.RawPayload, row.CreateTimes)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	existing, findErr := m.findByImmutableKey(ctx, row)
	if findErr == nil && sameAuthoritativeSnapshotIdentity(existing, row) {
		return nil
	}
	if findErr != nil && !errors.Is(findErr, ErrNotFound) {
		return findErr
	}
	existing, findErr = m.FindOneBySnapshotId(ctx, row.SnapshotId)
	if findErr == nil && sameAuthoritativeSnapshotIdentity(existing, row) {
		return nil
	}
	if findErr != nil && !errors.Is(findErr, ErrNotFound) {
		return findErr
	}
	return authoritativeSnapshotConflictError(row)
}

func authoritativeSnapshotConflictError(row *TItickAuthoritativeSnapshot) error {
	return fmt.Errorf("authoritative snapshot immutable-key conflict: authority=%s kind=%s symbol=%s source_timestamp=%d revision=%d", row.Authority, row.SnapshotKind, row.Symbol, row.SourceTimestamp, row.Revision)
}

func (m *customTItickAuthoritativeSnapshotModel) findByImmutableKey(ctx context.Context, row *TItickAuthoritativeSnapshot) (*TItickAuthoritativeSnapshot, error) {
	var existing TItickAuthoritativeSnapshot
	err := m.QueryRowNoCacheCtx(ctx, &existing, `SELECT id,snapshot_id,authority,snapshot_kind,category_code,market,symbol,price,source_timestamp,snapshot_timestamp,revision,formula_version,raw_payload,create_times
FROM t_itick_authoritative_snapshot WHERE authority=? AND snapshot_kind=? AND category_code=? AND market=? AND symbol=? AND source_timestamp=? AND revision=? LIMIT 1`, row.Authority, row.SnapshotKind, row.CategoryCode, row.Market, row.Symbol, row.SourceTimestamp, row.Revision)
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

func sameAuthoritativeSnapshotIdentity(a, b *TItickAuthoritativeSnapshot) bool {
	return a != nil && b != nil && a.SnapshotId == b.SnapshotId && a.Authority == b.Authority && a.SnapshotKind == b.SnapshotKind && a.CategoryCode == b.CategoryCode && a.Market == b.Market && a.Symbol == b.Symbol && a.Price.Equal(b.Price) && a.SourceTimestamp == b.SourceTimestamp && a.Revision == b.Revision && a.FormulaVersion == b.FormulaVersion
}

func (m *customTItickAuthoritativeSnapshotModel) FindAtOrBefore(ctx context.Context, authority, kind, category, market, symbol string, targetTime, minTime int64) (*TItickAuthoritativeSnapshot, error) {
	var row TItickAuthoritativeSnapshot
	err := m.QueryRowNoCacheCtx(ctx, &row, `SELECT id,snapshot_id,authority,snapshot_kind,category_code,market,symbol,price,source_timestamp,snapshot_timestamp,revision,formula_version,raw_payload,create_times
FROM t_itick_authoritative_snapshot
WHERE authority=? AND snapshot_kind=? AND category_code=? AND market=? AND symbol=? AND source_timestamp<=? AND source_timestamp>=?
	AND NOT EXISTS (SELECT 1 FROM t_itick_snapshot_revocation r WHERE r.snapshot_id=t_itick_authoritative_snapshot.snapshot_id)
ORDER BY source_timestamp DESC,revision DESC,id DESC LIMIT 1`, authority, kind, category, market, symbol, targetTime, minTime)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
