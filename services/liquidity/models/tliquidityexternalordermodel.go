package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityExternalOrderModel = (*customTLiquidityExternalOrderModel)(nil)

type (
	LiquidityExternalOrderPageFilter struct {
		TenantId, ProviderId, ConfigId, SymbolId int64
		Purpose, Side, Status                    int64
		Keyword                                  string
		TimeStart, TimeEnd                       int64
	}
	// TLiquidityExternalOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityExternalOrderModel.
	TLiquidityExternalOrderModel interface {
		tLiquidityExternalOrderModel
		FindPage(ctx context.Context, filter LiquidityExternalOrderPageFilter, cursor, limit int64) ([]*TLiquidityExternalOrder, int64, error)
	}

	customTLiquidityExternalOrderModel struct {
		*defaultTLiquidityExternalOrderModel
	}
)

// NewTLiquidityExternalOrderModel returns a model for the database table.
func NewTLiquidityExternalOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityExternalOrderModel {
	return &customTLiquidityExternalOrderModel{
		defaultTLiquidityExternalOrderModel: newTLiquidityExternalOrderModel(conn, c, opts...),
	}
}

func (m *customTLiquidityExternalOrderModel) FindPage(ctx context.Context, filter LiquidityExternalOrderPageFilter, cursor, limit int64) ([]*TLiquidityExternalOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", filter.TenantId)
	b.EqInt64("provider_id", filter.ProviderId)
	b.EqInt64("config_id", filter.ConfigId)
	b.EqInt64("symbol_id", filter.SymbolId)
	b.EqInt64("purpose", filter.Purpose)
	b.EqInt64("side", filter.Side)
	b.EqInt64("status", filter.Status)
	b.GteInt64("create_times", filter.TimeStart)
	b.LteInt64("create_times", filter.TimeEnd)
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		b.And("(order_no LIKE ? OR request_no LIKE ? OR external_order_id LIKE ?)", kw, kw, kw)
	}
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityExternalOrderRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityExternalOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
