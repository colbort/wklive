package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickSnapshotOutboxModel = (*customTItickSnapshotOutboxModel)(nil)

type (
	// TItickSnapshotOutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickSnapshotOutboxModel.
	TItickSnapshotOutboxModel interface {
		tItickSnapshotOutboxModel
		FindPending(context.Context, int64, int64) ([]*TItickSnapshotOutbox, error)
		Claim(context.Context, int64, int64) (bool, error)
		MarkSuccess(context.Context, int64, int64) error
		MarkFailure(context.Context, int64, string, int64) error
		FindPage(context.Context, int64, string, int64, int64) ([]*TItickSnapshotOutbox, int64, error)
		RetryFailed(context.Context, int64, int64) error
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

func (m *defaultTItickSnapshotOutboxModel) FindPage(ctx context.Context, status int64, snapshotID string, cursor, limit int64) ([]*TItickSnapshotOutbox, int64, error) {
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
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, "SELECT COUNT(1) FROM t_itick_snapshot_outbox WHERE "+countWhere, countArgs...); err != nil {
		return nil, 0, err
	}
	args = append(args, limit)
	var rows []*TItickSnapshotOutbox
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tItickSnapshotOutboxRows+" FROM t_itick_snapshot_outbox WHERE "+where+" ORDER BY id LIMIT ?", args...)
	return rows, total, err
}

func (m *defaultTItickSnapshotOutboxModel) RetryFailed(ctx context.Context, id, now int64) error {
	row, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	result, err := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET status=1,next_retry_at=?,last_error_msg='',update_times=? WHERE id=? AND status IN (4,5)", now, now, id)
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

func (m *defaultTItickSnapshotOutboxModel) FindPending(ctx context.Context, now, limit int64) ([]*TItickSnapshotOutbox, error) {
	var rows []*TItickSnapshotOutbox
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tItickSnapshotOutboxRows+" FROM t_itick_snapshot_outbox WHERE ((status IN (1,4) AND next_retry_at<=?) OR (status=2 AND update_times<=?)) ORDER BY id LIMIT ?", now, now-60000, limit)
	return rows, err
}
func (m *defaultTItickSnapshotOutboxModel) Claim(ctx context.Context, id, now int64) (bool, error) {
	r, e := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET status=2,update_times=? WHERE id=? AND ((status IN (1,4) AND next_retry_at<=?) OR (status=2 AND update_times<=?))", now, id, now, now-60000)
	if e != nil {
		return false, e
	}
	n, e := r.RowsAffected()
	return n == 1, e
}
func (m *defaultTItickSnapshotOutboxModel) MarkSuccess(ctx context.Context, id, now int64) error {
	_, e := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET status=3,next_retry_at=0,last_error_msg='',update_times=? WHERE id=? AND status=2", now, id)
	return e
}
func (m *defaultTItickSnapshotOutboxModel) MarkFailure(ctx context.Context, id int64, msg string, now int64) error {
	var retries int64
	if e := m.QueryRowNoCacheCtx(ctx, &retries, "SELECT retry_count FROM t_itick_snapshot_outbox WHERE id=?", id); e != nil {
		return e
	}
	retries++
	status, next := int64(4), now+(int64(1)<<min64(retries, 10))*int64(time.Second/time.Millisecond)
	if retries >= 20 {
		status, next = 5, 0
	}
	r, e := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_snapshot_outbox SET status=?,retry_count=?,next_retry_at=?,last_error_msg=?,update_times=? WHERE id=? AND status=2", status, retries, next, msg, now, id)
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
