package models

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickProductCalendarModel = (*customTItickProductCalendarModel)(nil)

type (
	// TItickProductCalendarModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickProductCalendarModel.
	TItickProductCalendarModel interface {
		tItickProductCalendarModel
		ResolveCalendar(ctx context.Context, category, market, symbol string) (*TItickMarketCalendar, error)
	}

	customTItickProductCalendarModel struct {
		*defaultTItickProductCalendarModel
	}
)

// NewTItickProductCalendarModel returns a model for the database table.
func NewTItickProductCalendarModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickProductCalendarModel {
	return &customTItickProductCalendarModel{
		defaultTItickProductCalendarModel: newTItickProductCalendarModel(conn, c, opts...),
	}
}

// ResolveCalendar returns the enabled calendar explicitly assigned to a
// product identity. Natural keys make mappings independent of product import
// order and database-generated IDs.
func (m *customTItickProductCalendarModel) ResolveCalendar(ctx context.Context, category, market, symbol string) (*TItickMarketCalendar, error) {
	var out TItickMarketCalendar
	err := m.QueryRowNoCacheCtx(ctx, &out, `SELECT c.id,c.category_code,c.market,c.exchange,c.timezone,
		c.trading_day_offset,c.week_start,c.enabled,c.remark,c.create_times,c.update_times
		FROM t_itick_product_calendar pc
		JOIN t_itick_market_calendar c ON c.id=pc.calendar_id
		WHERE pc.category_code=? AND pc.market=? AND pc.symbol=? AND c.enabled=1 LIMIT 1`,
		strings.ToLower(strings.TrimSpace(category)), strings.ToUpper(strings.TrimSpace(market)), strings.ToUpper(strings.TrimSpace(symbol)))
	if err != nil {
		return nil, err
	}
	return &out, nil
}
