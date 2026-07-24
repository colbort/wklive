package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityQuoteCycleModel = (*customTLiquidityQuoteCycleModel)(nil)

type (
	LiquidityQuoteCyclePageFilter struct {
		ConfigId, SymbolId, Status int64
		TimeStart, TimeEnd         int64
	}
	// TLiquidityQuoteCycleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityQuoteCycleModel.
	TLiquidityQuoteCycleModel interface {
		tLiquidityQuoteCycleModel
		FindPage(ctx context.Context, filter LiquidityQuoteCyclePageFilter, cursor, limit int64) ([]*TLiquidityQuoteCycle, int64, error)
	}

	customTLiquidityQuoteCycleModel struct {
		*defaultTLiquidityQuoteCycleModel
	}
)

// NewTLiquidityQuoteCycleModel returns a model for the database table.
func NewTLiquidityQuoteCycleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityQuoteCycleModel {
	return &customTLiquidityQuoteCycleModel{
		defaultTLiquidityQuoteCycleModel: newTLiquidityQuoteCycleModel(conn, c, opts...),
	}
}

func (m *customTLiquidityQuoteCycleModel) FindPage(ctx context.Context, filter LiquidityQuoteCyclePageFilter, cursor, limit int64) ([]*TLiquidityQuoteCycle, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("config_id", filter.ConfigId)
	b.EqInt64("symbol_id", filter.SymbolId)
	b.EqInt64("status", filter.Status)
	b.GteInt64("create_times", filter.TimeStart)
	b.LteInt64("create_times", filter.TimeEnd)
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityQuoteCycleRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityQuoteCycle
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
