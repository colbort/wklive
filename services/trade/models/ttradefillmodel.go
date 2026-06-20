package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTradeFillModel = (*customTTradeFillModel)(nil)

type (
	TradeFillPageFilter struct {
		TenantId   int64
		UserId     int64
		SymbolId   int64
		MarketType int64
		TimeStart  int64
		TimeEnd    int64
	}

	// TTradeFillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeFillModel.
	TTradeFillModel interface {
		tTradeFillModel
		FindPage(ctx context.Context, filter TradeFillPageFilter, cursor int64, limit int64) ([]*TTradeFill, int64, error)
		FindLastPrice(ctx context.Context, tenantId, symbolId, marketType int64) (float64, error)
	}

	customTTradeFillModel struct {
		*defaultTTradeFillModel
	}
)

// NewTTradeFillModel returns a model for the database table.
func NewTTradeFillModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeFillModel {
	return &customTTradeFillModel{
		defaultTTradeFillModel: newTTradeFillModel(conn, c, opts...),
	}
}

func (m *defaultTTradeFillModel) FindPage(ctx context.Context, filter TradeFillPageFilter, cursor int64, limit int64) ([]*TTradeFill, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.GteInt64("create_times", filter.TimeStart)
	builder.LteInt64("create_times", filter.TimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeFillRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTradeFill
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *defaultTTradeFillModel) FindLastPrice(ctx context.Context, tenantId, symbolId, marketType int64) (float64, error) {
	var price float64
	sql := fmt.Sprintf("SELECT `price` FROM %s WHERE `tenant_id` = ? AND `symbol_id` = ? AND `market_type` = ? ORDER BY `match_time` DESC, `id` DESC LIMIT 1", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &price, sql, tenantId, symbolId, marketType)
	switch err {
	case nil:
		return price, nil
	case sqlc.ErrNotFound:
		return 0, ErrNotFound
	default:
		return 0, err
	}
}
