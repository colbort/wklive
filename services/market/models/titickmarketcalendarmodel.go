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

var _ TItickMarketCalendarModel = (*customTMarketMarketCalendarModel)(nil)

type (
	// TItickMarketCalendarModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTMarketMarketCalendarModel.
	TItickMarketCalendarModel interface {
		tItickMarketCalendarModel
		Resolve(ctx context.Context, category, market, exchange string) (*TItickMarketCalendar, error)
		FindSessions(ctx context.Context, calendarID int64) ([]*TItickMarketSession, error)
		FindHoliday(ctx context.Context, calendarID int64, date time.Time) (*TItickMarketHoliday, error)
		Ensure(ctx context.Context, category, market, exchange, timezone string, now int64) (*TItickMarketCalendar, error)
	}

	customTMarketMarketCalendarModel struct {
		*defaultTMarketMarketCalendarModel
	}
)

// NewTMarketMarketCalendarModel returns a model for the database table.
func NewTMarketMarketCalendarModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickMarketCalendarModel {
	return &customTMarketMarketCalendarModel{
		defaultTMarketMarketCalendarModel: newTMarketMarketCalendarModel(conn, c, opts...),
	}
}

func (m *defaultTMarketMarketCalendarModel) Ensure(ctx context.Context, category, market, exchange, timezone string, now int64) (*TItickMarketCalendar, error) {
	category = strings.ToLower(strings.TrimSpace(category))
	market = strings.ToUpper(strings.TrimSpace(market))
	exchange = strings.TrimSpace(exchange)
	if timezone = strings.TrimSpace(timezone); timezone == "" {
		timezone = "UTC"
	}
	query := `INSERT INTO ` + m.table + `
		(category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
		VALUES (?,?,?,?,0,1,1,'iTick holiday sync',?,?)
		ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,update_times=VALUES(update_times)`
	if _, err := m.ExecNoCacheCtx(ctx, query, category, market, exchange, timezone, now, now); err != nil {
		return nil, err
	}
	var out TItickMarketCalendar
	if err := m.QueryRowNoCacheCtx(ctx, &out, `SELECT `+tItickMarketCalendarRows+` FROM `+m.table+`
		WHERE category_code=? AND market=? AND exchange=? LIMIT 1`, category, market, exchange); err != nil {
		return nil, err
	}
	return &out, nil
}

// Resolve uses the exact exchange first and falls back to the market default
// row (exchange=”). This query intentionally bypasses generated unique-key
// cache because it can match two candidate keys.
func (m *defaultTMarketMarketCalendarModel) Resolve(ctx context.Context, category, market, exchange string) (*TItickMarketCalendar, error) {
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

func (m *defaultTMarketMarketCalendarModel) FindSessions(ctx context.Context, calendarID int64) ([]*TItickMarketSession, error) {
	var out []*TItickMarketSession
	err := m.QueryRowsNoCacheCtx(ctx, &out, `SELECT `+tItickMarketSessionRows+`
		FROM t_itick_market_session WHERE calendar_id=? ORDER BY sort,id`, calendarID)
	return out, err
}

func (m *defaultTMarketMarketCalendarModel) FindHoliday(ctx context.Context, calendarID int64, date time.Time) (*TItickMarketHoliday, error) {
	var out TItickMarketHoliday
	err := m.QueryRowNoCacheCtx(ctx, &out, `SELECT `+tItickMarketHolidayRows+`
		FROM t_itick_market_holiday WHERE calendar_id=? AND trade_date=? LIMIT 1`, calendarID, date.Format("2006-01-02"))
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sqlx.ErrNotFound) {
		return nil, nil
	}
	return &out, err
}
