package models

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeOrderSecondsModel = (*customTTradeOrderSecondsModel)(nil)

type (
	SecondsOrderWorkItem struct {
		TTradeOrderSeconds
		OrderNo     string `db:"order_no"`
		UserId      int64  `db:"user_id"`
		SymbolId    int64  `db:"symbol_id"`
		OrderStatus int64  `db:"order_status"`
	}
	// TTradeOrderSecondsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeOrderSecondsModel.
	TTradeOrderSecondsModel interface {
		tTradeOrderSecondsModel
		FindWork(ctx context.Context, tenantID, status, dueAt, cursor, limit int64) ([]*SecondsOrderWorkItem, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TTradeOrderSeconds, error)
		SumExposure(ctx context.Context, tenantID, symbolID int64, statuses []int64) (decimal.Decimal, error)
	}

	customTTradeOrderSecondsModel struct {
		*defaultTTradeOrderSecondsModel
	}
)

// NewTTradeOrderSecondsModel returns a model for the database table.
func NewTTradeOrderSecondsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeOrderSecondsModel {
	return &customTTradeOrderSecondsModel{
		defaultTTradeOrderSecondsModel: newTTradeOrderSecondsModel(conn, c, opts...),
	}
}

func (m *defaultTTradeOrderSecondsModel) FindOneForUpdate(ctx context.Context, id int64) (*TTradeOrderSeconds, error) {
	var row TTradeOrderSeconds
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? FOR UPDATE", tTradeOrderSecondsRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, id); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *defaultTTradeOrderSecondsModel) FindWork(ctx context.Context, tenantID, status, dueAt, cursor, limit int64) ([]*SecondsOrderWorkItem, error) {
	where := "s.settlement_status = ? AND s.id > ?"
	args := []any{status, cursor}
	if tenantID > 0 {
		where += " AND s.tenant_id = ?"
		args = append(args, tenantID)
	}
	if dueAt > 0 {
		where += " AND s.expire_time <= ?"
		args = append(args, dueAt)
	}
	args = append(args, limit)
	query := fmt.Sprintf("SELECT %s, o.order_no, o.user_id, o.symbol_id, o.status AS order_status FROM %s s JOIN t_trade_order o ON o.id=s.order_id AND o.tenant_id=s.tenant_id WHERE %s ORDER BY s.id ASC LIMIT ?", prefixedFields(tTradeOrderSecondsFieldNames, "s"), m.table, where)
	var rows []*SecondsOrderWorkItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *defaultTTradeOrderSecondsModel) SumExposure(ctx context.Context, tenantID, symbolID int64, statuses []int64) (decimal.Decimal, error) {
	if len(statuses) == 0 {
		return decimal.Zero, nil
	}
	holders := "?"
	args := []any{tenantID, symbolID, statuses[0]}
	for _, status := range statuses[1:] {
		holders += ",?"
		args = append(args, status)
	}
	var amount decimal.Decimal
	query := fmt.Sprintf("SELECT COALESCE(SUM(s.stake_amount),0) FROM %s s JOIN t_trade_order o ON o.id=s.order_id WHERE s.tenant_id=? AND o.symbol_id=? AND s.settlement_status IN (%s)", m.table, holders)
	if err := m.QueryRowNoCacheCtx(ctx, &amount, query, args...); err != nil {
		return decimal.Zero, err
	}
	return amount, nil
}

func prefixedFields(fields []string, alias string) string {
	result := ""
	for i, field := range fields {
		if i > 0 {
			result += ","
		}
		result += alias + "." + field
	}
	return result
}
