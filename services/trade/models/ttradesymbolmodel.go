package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTradeSymbolModel = (*customTTradeSymbolModel)(nil)

type (
	TradeSymbolPageFilter struct {
		TenantId   int64
		MarketType int64
		Status     int64
		Keyword    string
	}

	// TTradeSymbolModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSymbolModel.
	TTradeSymbolModel interface {
		tTradeSymbolModel
		FindPage(ctx context.Context, filter TradeSymbolPageFilter, cursor int64, limit int64) ([]*TTradeSymbol, int64, error)
		FindAll(ctx context.Context, filter TradeSymbolPageFilter) ([]*TTradeSymbol, error)
	}

	customTTradeSymbolModel struct {
		*defaultTTradeSymbolModel
	}
)

// NewTTradeSymbolModel returns a model for the database table.
func NewTTradeSymbolModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolModel {
	return &customTTradeSymbolModel{
		defaultTTradeSymbolModel: newTTradeSymbolModel(conn, c, opts...),
	}
}

func (m *defaultTTradeSymbolModel) FindPage(ctx context.Context, filter TradeSymbolPageFilter, cursor int64, limit int64) ([]*TTradeSymbol, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqInt64("status", filter.Status)
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		builder.And("(symbol LIKE ? OR display_symbol LIKE ? OR base_asset LIKE ? OR quote_asset LIKE ? OR settle_asset LIKE ?)", kw, kw, kw, kw, kw)
	}

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeSymbolRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY sort ASC, id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTradeSymbol
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *defaultTTradeSymbolModel) FindAll(ctx context.Context, filter TradeSymbolPageFilter) ([]*TTradeSymbol, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqInt64("status", filter.Status)
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		builder.And("(symbol LIKE ? OR display_symbol LIKE ? OR base_asset LIKE ? OR quote_asset LIKE ? OR settle_asset LIKE ?)", kw, kw, kw, kw, kw)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY sort ASC, id DESC", tTradeSymbolRows, m.table, builder.Where())
	var list []*TTradeSymbol
	if err := m.QueryRowsNoCacheCtx(ctx, &list, sql, builder.Args()...); err != nil {
		return nil, err
	}
	return list, nil
}
