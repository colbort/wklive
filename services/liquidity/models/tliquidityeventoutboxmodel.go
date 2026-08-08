package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityEventOutboxModel = (*customTLiquidityEventOutboxModel)(nil)

type (
	// TLiquidityEventOutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityEventOutboxModel.
	TLiquidityEventOutboxModel interface {
		tLiquidityEventOutboxModel
		FindRunnable(ctx context.Context, now, staleBefore, limit int64) ([]*TLiquidityEventOutbox, error)
		Claim(ctx context.Context, id int64, owner string, now, staleBefore int64) (bool, error)
		MarkSuccess(ctx context.Context, id int64, owner string, now int64) (bool, error)
		MarkFailed(ctx context.Context, item *TLiquidityEventOutbox, owner, message string, now int64) (bool, error)
	}

	customTLiquidityEventOutboxModel struct {
		*defaultTLiquidityEventOutboxModel
	}
)

func (m *customTLiquidityEventOutboxModel) FindRunnable(
	ctx context.Context, now, staleBefore, limit int64,
) ([]*TLiquidityEventOutbox, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var rows []*TLiquidityEventOutbox
	err := m.QueryRowsNoCacheCtx(ctx, &rows, fmt.Sprintf(`SELECT %s FROM %s
WHERE ((status IN (1,4) AND next_retry_at<=?) OR (status=2 AND claimed_at<=?))
  AND (max_retry_count=0 OR retry_count<max_retry_count)
ORDER BY id LIMIT ?`, tLiquidityEventOutboxRows, m.table), now, staleBefore, limit)
	return rows, err
}

func (m *customTLiquidityEventOutboxModel) Claim(
	ctx context.Context, id int64, owner string, now, staleBefore int64,
) (bool, error) {
	if owner == "" {
		return false, fmt.Errorf("liquidity outbox lease owner is required")
	}
	result, err := m.ExecNoCacheCtx(ctx, fmt.Sprintf(`UPDATE %s
SET status=2,claimed_by=?,claimed_at=?,update_times=?
WHERE id=? AND ((status IN (1,4) AND next_retry_at<=?) OR (status=2 AND claimed_at<=?))
  AND (max_retry_count=0 OR retry_count<max_retry_count)`, m.table),
		owner, now, now, id, now, staleBefore,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func (m *customTLiquidityEventOutboxModel) MarkSuccess(
	ctx context.Context, id int64, owner string, now int64,
) (bool, error) {
	result, err := m.ExecNoCacheCtx(ctx, fmt.Sprintf(`UPDATE %s
SET status=3,sent_at=?,next_retry_at=0,last_error_msg='',claimed_by='',claimed_at=0,update_times=?
WHERE id=? AND status=2 AND claimed_by=?`, m.table), now, now, id, owner)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func (m *customTLiquidityEventOutboxModel) MarkFailed(
	ctx context.Context, item *TLiquidityEventOutbox, owner, message string, now int64,
) (bool, error) {
	if item == nil {
		return false, sql.ErrNoRows
	}
	retryCount := item.RetryCount + 1
	nextRetryAt := now + liquidityOutboxRetryDelay(retryCount).Milliseconds()
	if item.MaxRetryCount > 0 && retryCount >= item.MaxRetryCount {
		nextRetryAt = 0
	}
	runes := []rune(message)
	if len(runes) > 1000 {
		message = string(runes[:1000])
	}
	result, err := m.ExecNoCacheCtx(ctx, fmt.Sprintf(`UPDATE %s
SET status=4,retry_count=?,next_retry_at=?,last_error_msg=?,claimed_by='',claimed_at=0,update_times=?
WHERE id=? AND status=2 AND claimed_by=?`, m.table),
		retryCount, nextRetryAt, message, now, item.Id, owner,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func liquidityOutboxRetryDelay(retryCount int64) time.Duration {
	if retryCount < 1 {
		retryCount = 1
	}
	if retryCount > 6 {
		retryCount = 6
	}
	return time.Second * time.Duration(1<<(retryCount-1))
}

// NewTLiquidityEventOutboxModel returns a model for the database table.
func NewTLiquidityEventOutboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityEventOutboxModel {
	return &customTLiquidityEventOutboxModel{
		defaultTLiquidityEventOutboxModel: newTLiquidityEventOutboxModel(conn, c, opts...),
	}
}
