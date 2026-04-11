package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type TradeOrderPageFilter struct {
	TenantId        int64
	UserId          int64
	SymbolId        int64
	MarketType      int64
	Status          int64
	Side            int64
	TimeStart       int64
	TimeEnd         int64
	Keyword         string
	Statuses        []int64
	ExcludeStatuses []int64
	PositionSide    int64
}

type TradeOrderModel interface {
	tTradeOrderModel
	FindPage(ctx context.Context, filter TradeOrderPageFilter, cursor int64, limit int64) ([]*TTradeOrder, int64, error)
	CountByStatuses(ctx context.Context, tenantId, userId uint64, marketType int64, statuses []int64) (int64, error)
}

func (m *defaultTTradeOrderModel) FindPage(ctx context.Context, filter TradeOrderPageFilter, cursor int64, limit int64) ([]*TTradeOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("side", filter.Side)
	builder.EqInt64("position_side", filter.PositionSide)
	builder.GteInt64("create_times", filter.TimeStart)
	builder.LteInt64("create_times", filter.TimeEnd)
	builder.InInt64("status", filter.Statuses)
	if len(filter.ExcludeStatuses) > 0 {
		holders := make([]any, 0, len(filter.ExcludeStatuses))
		parts := make([]string, 0, len(filter.ExcludeStatuses))
		for _, item := range filter.ExcludeStatuses {
			parts = append(parts, "?")
			holders = append(holders, item)
		}
		builder.And(fmt.Sprintf("status NOT IN (%s)", joinComma(parts)), holders...)
	}
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		builder.And("(order_no LIKE ? OR client_order_id LIKE ?)", kw, kw)
	}

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeOrderRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTradeOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *defaultTTradeOrderModel) CountByStatuses(ctx context.Context, tenantId, userId uint64, marketType int64, statuses []int64) (int64, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", int64(tenantId))
	builder.EqInt64("user_id", int64(userId))
	builder.EqInt64("market_type", marketType)
	builder.InInt64("status", statuses)

	var total int64
	sql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, builder.Where())
	if err := m.QueryRowNoCacheCtx(ctx, &total, sql, builder.Args()...); err != nil {
		return 0, err
	}
	return total, nil
}

func joinComma(items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for i := 1; i < len(items); i++ {
		out += "," + items[i]
	}
	return out
}
