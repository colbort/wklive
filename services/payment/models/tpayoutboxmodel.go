package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TPayOutboxModel = (*customTPayOutboxModel)(nil)

type (
	// TPayOutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPayOutboxModel.
	TPayOutboxModel interface {
		tPayOutboxModel
		FindPending(ctx context.Context, now int64, limit int64) ([]*TPayOutbox, error)
		ClaimPending(ctx context.Context, workerID string, now int64, limit int64) ([]*TPayOutbox, error)
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

func (m *customTPayOutboxModel) FindPending(ctx context.Context, now int64, limit int64) ([]*TPayOutbox, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []*TPayOutbox
	query := fmt.Sprintf("select %s from %s where `status` in (1,4) and `next_retry_at` <= ? order by `id` asc limit ?", tPayOutboxRows, m.table)
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, now, limit); err != nil {
		return nil, err
	}
	return rows, nil
}

// ClaimPending implements [TPayOutboxModel].
func (m *customTPayOutboxModel) ClaimPending(ctx context.Context, workerID string, now int64, limit int64) ([]*TPayOutbox, error) {
	panic("unimplemented")
}
