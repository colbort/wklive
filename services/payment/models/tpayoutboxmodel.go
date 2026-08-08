package models

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TPayOutboxModel = (*customTPayOutboxModel)(nil)

type (
	// TPayOutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPayOutboxModel.
	TPayOutboxModel interface {
		tPayOutboxModel
		ClaimPending(ctx context.Context, owner string, now, staleBefore, limit int64) ([]*TPayOutbox, error)
		MarkSuccess(ctx context.Context, row *TPayOutbox, owner string, now int64) (bool, error)
		MarkFailed(ctx context.Context, row *TPayOutbox, owner string, retryCount, nextRetryAt, now int64, message string) (bool, error)
	}

	customTPayOutboxModel struct {
		*defaultTPayOutboxModel
	}
)

// NewTPayOutboxModel returns a model for the database table.
func NewTPayOutboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TPayOutboxModel {
	return &customTPayOutboxModel{
		defaultTPayOutboxModel: newTPayOutboxModel(conn, c, opts...),
	}
}

// ClaimPending uses a compare-and-set update for every candidate. Multiple
// Payment replicas may observe the same candidate, but only one can move it to
// PROCESSING with its worker ID. Expired PROCESSING leases are recoverable.
func (m *customTPayOutboxModel) ClaimPending(ctx context.Context, owner string, now, staleBefore, limit int64) ([]*TPayOutbox, error) {
	if owner == "" {
		return nil, fmt.Errorf("payment outbox owner must not be empty")
	}
	if limit <= 0 {
		limit = 100
	}
	var candidates []*TPayOutbox
	query := fmt.Sprintf(
		"select %s from %s where ((`status` in (1,4) and `next_retry_at` <= ?) or (`status` = 2 and `claimed_at` <= ?)) order by `id` asc limit ?",
		tPayOutboxRows,
		m.table,
	)
	if err := m.QueryRowsNoCacheCtx(ctx, &candidates, query, now, staleBefore, limit); err != nil {
		return nil, err
	}

	claimed := make([]*TPayOutbox, 0, len(candidates))
	for _, row := range candidates {
		idKey := fmt.Sprintf("%s%v", cacheTPayOutboxIdPrefix, row.Id)
		eventKey := fmt.Sprintf("%s%v", cacheTPayOutboxEventNoPrefix, row.EventNo)
		result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
			claimSQL := fmt.Sprintf(`update %s
set status = ?, claimed_by = ?, claimed_at = ?, update_times = ?
where id = ?
  and ((status in (?, ?) and next_retry_at <= ?)
       or (status = ? and claimed_at <= ?))`, m.table)
			return conn.ExecCtx(
				ctx,
				claimSQL,
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PROCESSING),
				owner,
				now,
				now,
				row.Id,
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PENDING),
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_FAILED),
				now,
				int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PROCESSING),
				staleBefore,
			)
		}, idKey, eventKey)
		if err != nil {
			return claimed, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return claimed, err
		}
		if affected != 1 {
			continue
		}
		row.Status = int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PROCESSING)
		row.ClaimedBy = owner
		row.ClaimedAt = now
		row.UpdateTimes = now
		claimed = append(claimed, row)
	}
	return claimed, nil
}

func (m *customTPayOutboxModel) MarkSuccess(ctx context.Context, row *TPayOutbox, owner string, now int64) (bool, error) {
	return m.updateClaimed(
		ctx,
		row,
		owner,
		"status = ?, claimed_by = '', claimed_at = 0, next_retry_at = 0, last_error_msg = '', update_times = ?",
		int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_SUCCESS),
		now,
	)
}

func (m *customTPayOutboxModel) MarkFailed(
	ctx context.Context,
	row *TPayOutbox,
	owner string,
	retryCount, nextRetryAt, now int64,
	message string,
) (bool, error) {
	return m.updateClaimed(
		ctx,
		row,
		owner,
		"status = ?, retry_count = ?, next_retry_at = ?, last_error_msg = ?, claimed_by = '', claimed_at = 0, update_times = ?",
		int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_FAILED),
		retryCount,
		nextRetryAt,
		truncatePayOutboxError(message),
		now,
	)
}

func (m *customTPayOutboxModel) updateClaimed(
	ctx context.Context,
	row *TPayOutbox,
	owner, setClause string,
	setArgs ...any,
) (bool, error) {
	if row == nil || owner == "" {
		return false, nil
	}
	idKey := fmt.Sprintf("%s%v", cacheTPayOutboxIdPrefix, row.Id)
	eventKey := fmt.Sprintf("%s%v", cacheTPayOutboxEventNoPrefix, row.EventNo)
	args := append(append([]any{}, setArgs...), row.Id, int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_PROCESSING), owner)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("update %s set %s where id = ? and status = ? and claimed_by = ?", m.table, setClause)
		return conn.ExecCtx(ctx, query, args...)
	}, idKey, eventKey)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func truncatePayOutboxError(message string) string {
	runes := []rune(message)
	if len(runes) <= 1000 {
		return message
	}
	return string(runes[:1000])
}
