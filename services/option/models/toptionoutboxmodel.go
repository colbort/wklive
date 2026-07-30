package models

import (
	"context"
	"fmt"

	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionOutboxModel = (*customTOptionOutboxModel)(nil)

type (
	// TOptionOutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionOutboxModel.
	TOptionOutboxModel interface {
		tOptionOutboxModel
		FindRunnable(ctx context.Context, tenantId, now, limit int64) ([]*TOptionOutbox, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionOutbox, error)
		Claim(ctx context.Context, id, now int64) (bool, error)
		RecoverStale(ctx context.Context, staleBefore, now int64) error
		HasIncomplete(ctx context.Context, tenantId, contractId int64) (bool, error)
		ResetForManualRetry(ctx context.Context, id, now int64) (bool, error)
	}

	customTOptionOutboxModel struct {
		*defaultTOptionOutboxModel
	}
)

func (m *defaultTOptionOutboxModel) FindRunnable(ctx context.Context, tenantId, now, limit int64) ([]*TOptionOutbox, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	query := fmt.Sprintf(`SELECT %s FROM %s AS current
WHERE (? = 0 OR current.tenant_id = ?)
AND current.status IN (?, ?) AND current.next_retry_at <= ?
AND NOT EXISTS (
  SELECT 1 FROM %s AS previous
  WHERE previous.tenant_id = current.tenant_id
    AND previous.contract_id = current.contract_id
    AND previous.event_type = current.event_type
    AND previous.match_sequence < current.match_sequence
    AND previous.status <> ?
)
ORDER BY current.id LIMIT ?`, tOptionOutboxRows, m.table, m.table)
	var list []*TOptionOutbox
	err := m.QueryRowsNoCacheCtx(ctx, &list, query,
		tenantId, tenantId,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED),
		now,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS),
		limit,
	)
	return list, err
}

func (m *defaultTOptionOutboxModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionOutbox, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionOutboxRows, m.table)
	var item TOptionOutbox
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionOutboxModel) Claim(ctx context.Context, id, now int64) (bool, error) {
	query := fmt.Sprintf("UPDATE %s SET status = ?, update_times = ? WHERE id = ? AND status IN (?, ?)", m.table)
	result, err := m.ExecNoCacheCtx(ctx, query,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PROCESSING), now, id,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED),
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func (m *defaultTOptionOutboxModel) RecoverStale(ctx context.Context, staleBefore, now int64) error {
	query := fmt.Sprintf(`UPDATE %s SET status = ?, next_retry_at = ?, update_times = ?,
last_error_msg = 'recovered stale processing event'
WHERE status = ? AND update_times < ?`, m.table)
	_, err := m.ExecNoCacheCtx(ctx, query,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED), now, now,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PROCESSING), staleBefore,
	)
	return err
}

func (m *defaultTOptionOutboxModel) HasIncomplete(ctx context.Context, tenantId, contractId int64) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id = ? AND contract_id = ? AND status <> ?", m.table)
	var count int64
	if err := m.QueryRowNoCacheCtx(ctx, &count, query, tenantId, contractId, int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS)); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *defaultTOptionOutboxModel) ResetForManualRetry(ctx context.Context, id, now int64) (bool, error) {
	query := fmt.Sprintf(`UPDATE %s SET status = ?, retry_count = 0, next_retry_at = ?,
last_error_msg = '', update_times = ? WHERE id = ? AND status IN (?, ?)`, m.table)
	result, err := m.ExecNoCacheCtx(ctx, query,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING), now, now, id,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_MANUAL_REVIEW),
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

// NewTOptionOutboxModel returns a model for the database table.
func NewTOptionOutboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionOutboxModel {
	return &customTOptionOutboxModel{
		defaultTOptionOutboxModel: newTOptionOutboxModel(conn, c, opts...),
	}
}
