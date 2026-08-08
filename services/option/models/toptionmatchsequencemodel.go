package models

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionMatchSequenceModel = (*customTOptionMatchSequenceModel)(nil)

type (
	// TOptionMatchSequenceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionMatchSequenceModel.
	TOptionMatchSequenceModel interface {
		tOptionMatchSequenceModel
		Next(ctx context.Context, tenantId, contractId int64) (int64, error)
	}

	customTOptionMatchSequenceModel struct {
		*defaultTOptionMatchSequenceModel
	}
)

func (m *customTOptionMatchSequenceModel) Next(ctx context.Context, tenantId, contractId int64) (int64, error) {
	now := time.Now().Unix()
	insert := fmt.Sprintf(`INSERT INTO %s (tenant_id, contract_id, next_sequence, update_times)
VALUES (?, ?, 1, ?) ON DUPLICATE KEY UPDATE update_times = update_times`, m.table)
	if _, err := m.ExecNoCacheCtx(ctx, insert, tenantId, contractId, now); err != nil {
		return 0, err
	}
	var next int64
	query := fmt.Sprintf("SELECT next_sequence FROM %s WHERE tenant_id = ? AND contract_id = ? FOR UPDATE", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &next, query, tenantId, contractId); err != nil {
		return 0, err
	}
	if next <= 0 {
		return 0, fmt.Errorf("invalid option match sequence: %d", next)
	}
	update := fmt.Sprintf("UPDATE %s SET next_sequence = ?, update_times = ? WHERE tenant_id = ? AND contract_id = ?", m.table)
	if _, err := m.ExecNoCacheCtx(ctx, update, next+1, now, tenantId, contractId); err != nil {
		return 0, err
	}
	return next, nil
}

// NewTOptionMatchSequenceModel returns a model for the database table.
func NewTOptionMatchSequenceModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionMatchSequenceModel {
	return &customTOptionMatchSequenceModel{
		defaultTOptionMatchSequenceModel: newTOptionMatchSequenceModel(conn, c, opts...),
	}
}
