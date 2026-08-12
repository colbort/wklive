package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickSnapshotOutboxModel = (*customTItickSnapshotOutboxModel)(nil)

type (
	SnapshotOutboxHealth struct {
		PendingCount    int64 `db:"pending_count"`
		ProcessingCount int64 `db:"processing_count"`
		FailedCount     int64 `db:"failed_count"`
		ManualCount     int64 `db:"manual_count"`
		OldestOpenAt    int64 `db:"oldest_open_at"`
	}

	// TItickSnapshotOutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickSnapshotOutboxModel.
	TItickSnapshotOutboxModel interface {
		tItickSnapshotOutboxModel
		FindPending(context.Context, int64, int64) ([]*TItickSnapshotOutbox, error)
		Claim(context.Context, int64, string, int64, int64) (bool, error)
		MarkSuccess(context.Context, int64, string, int64) error
		MarkFailure(context.Context, int64, string, string, int64) error
		MarkRedisPublished(context.Context, int64, string, int64) error
		MarkEventPublished(context.Context, int64, string, int64) error
		CompleteAfterEventPublished(context.Context, int64, string, int64) error
		FindPage(context.Context, int64, string, int64, int64, ...int64) ([]*TItickSnapshotOutbox, int64, error)
		RetryFailed(context.Context, int64, int64) error
		Health(context.Context) (*SnapshotOutboxHealth, error)
		DeleteSucceededBefore(context.Context, int64, int64) (int64, error)
	}

	customTItickSnapshotOutboxModel struct {
		*defaultTItickSnapshotOutboxModel
	}
)

// NewTItickSnapshotOutboxModel returns a model for the database table.
func NewTItickSnapshotOutboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickSnapshotOutboxModel {
	return &customTItickSnapshotOutboxModel{
		defaultTItickSnapshotOutboxModel: newTItickSnapshotOutboxModel(conn, c, opts...),
	}
}

func (m *customTItickSnapshotOutboxModel) DeleteSucceededBefore(ctx context.Context, cutoff, limit int64) (int64, error) {
	if limit <= 0 || limit > 10000 {
		limit = 5000
	}
	result, err := m.ExecNoCacheCtx(ctx, `DELETE FROM t_itick_snapshot_outbox
		WHERE status=3 AND update_times<? ORDER BY id LIMIT ?`, cutoff, limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (m *customTItickSnapshotOutboxModel) MarkRedisPublished(ctx context.Context, id int64, owner string, now int64) error {
	result, err := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET redis_published_at=CASE WHEN redis_published_at=0 THEN ? ELSE redis_published_at END,update_times=? WHERE id=? AND status=2 AND claimed_by=?", now, now, id, owner)
	return requireOneOutboxRow(result, err)
}

func (m *customTItickSnapshotOutboxModel) MarkEventPublished(ctx context.Context, id int64, owner string, now int64) error {
	result, err := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET event_published_at=CASE WHEN event_published_at=0 THEN ? ELSE event_published_at END,update_times=? WHERE id=? AND status=2 AND claimed_by=?", now, now, id, owner)
	return requireOneOutboxRow(result, err)
}

// CompleteAfterEventPublished atomically checkpoints the final publication
// stage and closes the outbox row. A retry can only reach this method after the
// Redis checkpoint is durable, so success never hides an incomplete Redis
// publication.
func (m *customTItickSnapshotOutboxModel) CompleteAfterEventPublished(ctx context.Context, id int64, owner string, now int64) error {
	result, err := m.ExecNoCacheCtx(ctx, `UPDATE t_itick_snapshot_outbox
		SET event_published_at=CASE WHEN event_published_at=0 THEN ? ELSE event_published_at END,
		    status=3,next_retry_at=0,last_error_msg='',claimed_by='',claimed_at=0,update_times=?
		WHERE id=? AND status=2 AND claimed_by=? AND redis_published_at>0`, now, now, id, owner)
	return requireOneOutboxRow(result, err)
}

func requireOneOutboxRow(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (m *customTItickSnapshotOutboxModel) Health(ctx context.Context) (*SnapshotOutboxHealth, error) {
	var rows []struct {
		Status   int64 `db:"status"`
		Count    int64 `db:"row_count"`
		OldestAt int64 `db:"oldest_at"`
	}
	query := `SELECT status,COUNT(*) AS row_count,MIN(create_times) AS oldest_at
		FROM t_itick_snapshot_outbox
		WHERE status IN (1,2,4,5)
		GROUP BY status`
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		return nil, err
	}
	health := SnapshotOutboxHealth{}
	for _, row := range rows {
		switch row.Status {
		case 1:
			health.PendingCount = row.Count
		case 2:
			health.ProcessingCount = row.Count
		case 4:
			health.FailedCount = row.Count
		case 5:
			health.ManualCount = row.Count
		}
		if health.OldestOpenAt == 0 || row.OldestAt < health.OldestOpenAt {
			health.OldestOpenAt = row.OldestAt
		}
	}
	return &health, nil
}

func (m *customTItickSnapshotOutboxModel) FindPage(ctx context.Context, status int64, snapshotID string, cursor, limit int64, knownCounts ...int64) ([]*TItickSnapshotOutbox, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	where, args := "id>?", []any{cursor}
	countWhere, countArgs := "id>0", []any{}
	if status > 0 {
		where, countWhere = where+" AND status=?", countWhere+" AND status=?"
		args, countArgs = append(args, status), append(countArgs, status)
	}
	if snapshotID != "" {
		where, countWhere = where+" AND snapshot_id=?", countWhere+" AND snapshot_id=?"
		args, countArgs = append(args, snapshotID), append(countArgs, snapshotID)
	}
	total := sqlutil.KnownCount(knownCounts...)
	if total <= 0 {
		if err := m.QueryRowNoCacheCtx(ctx, &total, "SELECT COUNT(1) FROM t_itick_snapshot_outbox WHERE "+countWhere, countArgs...); err != nil {
			return nil, 0, err
		}
	}
	args = append(args, limit)
	var rows []*TItickSnapshotOutbox
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tItickSnapshotOutboxRows+" FROM t_itick_snapshot_outbox WHERE "+where+" ORDER BY id LIMIT ?", args...)
	return rows, total, err
}

func (m *customTItickSnapshotOutboxModel) RetryFailed(ctx context.Context, id, now int64) error {
	row, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	result, err := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET status=1,next_retry_at=?,last_error_msg='',claimed_by='',claimed_at=0,update_times=? WHERE id=? AND status IN (4,5)", now, now, id)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return sql.ErrNoRows
	}
	return m.DelCacheCtx(ctx,
		fmt.Sprintf("%s%v", cacheTItickSnapshotOutboxIdPrefix, row.Id),
		fmt.Sprintf("%s%v", cacheTItickSnapshotOutboxSnapshotIdPrefix, row.SnapshotId),
	)
}

func (m *customTItickSnapshotOutboxModel) FindPending(ctx context.Context, now, limit int64) ([]*TItickSnapshotOutbox, error) {
	var rows []*TItickSnapshotOutbox
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tItickSnapshotOutboxRows+" FROM t_itick_snapshot_outbox WHERE ((status IN (1,4) AND next_retry_at<=?) OR (status=2 AND claimed_at<=?)) ORDER BY id LIMIT ?", now, now-60000, limit)
	return rows, err
}
func (m *customTItickSnapshotOutboxModel) Claim(ctx context.Context, id int64, owner string, now, staleBefore int64) (bool, error) {
	if owner == "" {
		return false, fmt.Errorf("snapshot outbox lease owner is required")
	}
	r, e := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET status=2,claimed_by=?,claimed_at=?,update_times=? WHERE id=? AND ((status IN (1,4) AND next_retry_at<=?) OR (status=2 AND claimed_at<=?))", owner, now, now, id, now, staleBefore)
	if e != nil {
		return false, e
	}
	n, e := r.RowsAffected()
	return n == 1, e
}
func (m *customTItickSnapshotOutboxModel) MarkSuccess(ctx context.Context, id int64, owner string, now int64) error {
	result, err := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET status=3,next_retry_at=0,last_error_msg='',claimed_by='',claimed_at=0,update_times=? WHERE id=? AND status=2 AND claimed_by=? AND redis_published_at>0 AND event_published_at>0", now, id, owner)
	return requireOneOutboxRow(result, err)
}
func (m *customTItickSnapshotOutboxModel) MarkFailure(ctx context.Context, id int64, owner, msg string, now int64) error {
	runes := []rune(msg)
	if len(runes) > 500 {
		msg = string(runes[:500])
	}
	var retries int64
	if e := m.QueryRowNoCacheCtx(ctx, &retries, "SELECT retry_count FROM t_itick_snapshot_outbox WHERE id=?", id); e != nil {
		return e
	}
	retries++
	status, next := int64(4), now+(int64(1)<<min64(retries, 10))*int64(time.Second/time.Millisecond)
	if retries >= 20 {
		status, next = 5, 0
	}
	r, e := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET status=?,retry_count=?,next_retry_at=?,last_error_msg=?,claimed_by='',claimed_at=0,update_times=? WHERE id=? AND status=2 AND claimed_by=?", status, retries, next, msg, now, id, owner)
	if e != nil {
		return e
	}
	n, e := r.RowsAffected()
	if n != 1 && e == nil {
		return sql.ErrNoRows
	}
	return e
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
