package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickMarketCalendarModel = (*customTItickMarketCalendarModel)(nil)

type (
	// TItickMarketCalendarModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickMarketCalendarModel.
	TItickMarketCalendarModel interface {
		tItickMarketCalendarModel
		Resolve(ctx context.Context, category, market, exchange string) (*TItickMarketCalendar, error)
		FindSessions(ctx context.Context, calendarID int64) ([]*TItickMarketSession, error)
		FindHoliday(ctx context.Context, calendarID int64, date time.Time) (*TItickMarketHoliday, error)
	}

	customTItickMarketCalendarModel struct {
		*defaultTItickMarketCalendarModel
	}
)

// NewTItickMarketCalendarModel returns a model for the database table.
func NewTItickMarketCalendarModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickMarketCalendarModel {
	return &customTItickMarketCalendarModel{
		defaultTItickMarketCalendarModel: newTItickMarketCalendarModel(conn, c, opts...),
	}
}

// Resolve uses the exact exchange first and falls back to the market default
// row (exchange=”). This query intentionally bypasses generated unique-key
// cache because it can match two candidate keys.
func (m *defaultTItickMarketCalendarModel) Resolve(ctx context.Context, category, market, exchange string) (*TItickMarketCalendar, error) {
	category = strings.ToLower(strings.TrimSpace(category))
	market = strings.ToUpper(strings.TrimSpace(market))
	exchange = strings.TrimSpace(exchange)
	var out TItickMarketCalendar
	query := `SELECT ` + tItickMarketCalendarRows + ` FROM ` + m.table + `
		WHERE category_code=? AND market=? AND exchange IN (?, '') AND enabled=1
		ORDER BY (exchange = ?) DESC LIMIT 1`
	if err := m.QueryRowNoCacheCtx(ctx, &out, query, category, market, exchange, exchange); err != nil {
		return nil, err
	}
	return &out, nil
}

func (m *defaultTItickMarketCalendarModel) FindSessions(ctx context.Context, calendarID int64) ([]*TItickMarketSession, error) {
	var out []*TItickMarketSession
	err := m.QueryRowsNoCacheCtx(ctx, &out, `SELECT `+tItickMarketSessionRows+`
		FROM t_itick_market_session WHERE calendar_id=? ORDER BY sort,id`, calendarID)
	return out, err
}

func (m *defaultTItickMarketCalendarModel) FindHoliday(ctx context.Context, calendarID int64, date time.Time) (*TItickMarketHoliday, error) {
	var out TItickMarketHoliday
	err := m.QueryRowNoCacheCtx(ctx, &out, `SELECT `+tItickMarketHolidayRows+`
		FROM t_itick_market_holiday WHERE calendar_id=? AND trade_date=? LIMIT 1`, calendarID, date.Format("2006-01-02"))
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sqlx.ErrNotFound) {
		return nil, nil
	}
	return &out, err
}
