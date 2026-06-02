package models

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
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

type TradeOrderMatchKey struct {
	TenantId   int64 `db:"tenant_id"`
	SymbolId   int64 `db:"symbol_id"`
	MarketType int64 `db:"market_type"`
}

type TradeOrderModel interface {
	tTradeOrderModel
	FindPage(ctx context.Context, filter TradeOrderPageFilter, cursor int64, limit int64) ([]*TTradeOrder, int64, error)
	CountByStatuses(ctx context.Context, tenantId, userId uint64, marketType int64, statuses []int64) (int64, error)
	FindMatchKeys(ctx context.Context, tenantId int64, statuses []int64, limit int64) ([]TradeOrderMatchKey, error)
	FindOpenMatchOrders(ctx context.Context, tenantId, symbolId, marketType, side int64, statuses []int64, marketOrderType int64, limit int64) ([]*TTradeOrder, error)
	FindOneForUpdate(ctx context.Context, id int64) (*TTradeOrder, error)
	FindOneByTenantIdOrderNoForUpdate(ctx context.Context, tenantId int64, orderNo string) (*TTradeOrder, error)
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

func (m *defaultTTradeOrderModel) FindMatchKeys(ctx context.Context, tenantId int64, statuses []int64, limit int64) ([]TradeOrderMatchKey, error) {
	limit = sqlutil.NormalizeLimit(limit)
	where, args := openOrderWhere(tenantId, 0, 0, 0, statuses)
	sql := fmt.Sprintf("SELECT tenant_id, symbol_id, market_type FROM %s WHERE %s AND order_type IN (?, ?) GROUP BY tenant_id, symbol_id, market_type ORDER BY tenant_id ASC, symbol_id ASC, market_type ASC LIMIT ?", m.table, where)
	args = append(args, 1, 2, limit)

	var list []TradeOrderMatchKey
	if err := m.QueryRowsNoCacheCtx(ctx, &list, sql, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTTradeOrderModel) FindOpenMatchOrders(ctx context.Context, tenantId, symbolId, marketType, side int64, statuses []int64, marketOrderType int64, limit int64) ([]*TTradeOrder, error) {
	limit = sqlutil.NormalizeLimit(limit)
	where, args := openOrderWhere(tenantId, symbolId, marketType, side, statuses)

	priceOrder := "price ASC"
	if side == 1 {
		priceOrder = "price DESC"
	}
	sql := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s AND order_type IN (?, ?) ORDER BY CASE WHEN order_type = ? THEN 0 ELSE 1 END ASC, %s, id ASC LIMIT ?",
		tTradeOrderRows,
		m.table,
		where,
		priceOrder,
	)
	args = append(args, 1, 2, marketOrderType, limit)

	var list []*TTradeOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, sql, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTTradeOrderModel) FindOneForUpdate(ctx context.Context, id int64) (*TTradeOrder, error) {
	var resp TTradeOrder
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? LIMIT 1 FOR UPDATE", tTradeOrderRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, sql, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultTTradeOrderModel) FindOneByTenantIdOrderNoForUpdate(ctx context.Context, tenantId int64, orderNo string) (*TTradeOrder, error) {
	var resp TTradeOrder
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE `tenant_id` = ? AND `order_no` = ? LIMIT 1 FOR UPDATE", tTradeOrderRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, sql, tenantId, orderNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func openOrderWhere(tenantId, symbolId, marketType, side int64, statuses []int64) (string, []any) {
	parts := make([]string, 0, 5)
	args := make([]any, 0, 8)
	if tenantId > 0 {
		parts = append(parts, "tenant_id = ?")
		args = append(args, tenantId)
	}
	if symbolId > 0 {
		parts = append(parts, "symbol_id = ?")
		args = append(args, symbolId)
	}
	if marketType > 0 {
		parts = append(parts, "market_type = ?")
		args = append(args, marketType)
	}
	if side > 0 {
		parts = append(parts, "side = ?")
		args = append(args, side)
	}
	if len(statuses) > 0 {
		holders := make([]string, 0, len(statuses))
		for _, status := range statuses {
			holders = append(holders, "?")
			args = append(args, status)
		}
		parts = append(parts, fmt.Sprintf("status IN (%s)", joinComma(holders)))
	}
	if len(parts) == 0 {
		return "1=1", args
	}
	return strings.Join(parts, " AND "), args
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
