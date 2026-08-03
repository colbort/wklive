package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeUserControlAuditModel = (*customTTradeUserControlAuditModel)(nil)

type (
	// TTradeUserControlAuditModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeUserControlAuditModel.
	TTradeUserControlAuditModel interface {
		tTradeUserControlAuditModel
		FindPage(ctx context.Context, filter UserTradeControlFilter, cursor int64, limit int64) ([]*TTradeUserControlAudit, int64, error)
	}

	customTTradeUserControlAuditModel struct {
		*defaultTTradeUserControlAuditModel
	}
)

// NewTTradeUserControlAuditModel returns a model for the database table.
func NewTTradeUserControlAuditModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeUserControlAuditModel {
	return &customTTradeUserControlAuditModel{
		defaultTTradeUserControlAuditModel: newTTradeUserControlAuditModel(conn, c, opts...),
	}
}

func (m *defaultTTradeUserControlAuditModel) FindPage(ctx context.Context, filter UserTradeControlFilter, cursor int64, limit int64) ([]*TTradeUserControlAudit, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", filter.TenantId)
	b.EqInt64("user_id", filter.UserId)
	b.EqInt64("scope_type", filter.ScopeType)
	b.EqInt64("control_id", filter.ControlId)
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeUserControlAuditRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var list []*TTradeUserControlAudit
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
