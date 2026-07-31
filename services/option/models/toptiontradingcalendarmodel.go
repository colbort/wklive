package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionTradingCalendarModel = (*customTOptionTradingCalendarModel)(nil)

type (
	OptionTradingCalendarPageFilter struct {
		TenantId     int64
		CalendarCode string
		Status       int64
	}

	// TOptionTradingCalendarModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionTradingCalendarModel.
	TOptionTradingCalendarModel interface {
		tOptionTradingCalendarModel
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionTradingCalendar, error)
		FindLatestForUpdate(ctx context.Context, tenantId int64, calendarCode string) (*TOptionTradingCalendar, error)
		FindOpenEndedApprovedForUpdate(ctx context.Context, tenantId int64, calendarCode string) (*TOptionTradingCalendar, error)
		FindEffective(ctx context.Context, tenantId int64, calendarCode string, now int64) (*TOptionTradingCalendar, error)
		FindPage(ctx context.Context, filter OptionTradingCalendarPageFilter, cursor, limit int64) ([]*TOptionTradingCalendar, int64, error)
	}

	customTOptionTradingCalendarModel struct {
		*defaultTOptionTradingCalendarModel
	}
)

func (m *defaultTOptionTradingCalendarModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionTradingCalendar, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1 FOR UPDATE",
		tOptionTradingCalendarRows, m.table)
	var item TOptionTradingCalendar
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionTradingCalendarModel) FindLatestForUpdate(
	ctx context.Context, tenantId int64, calendarCode string,
) (*TOptionTradingCalendar, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND calendar_code=? ORDER BY version DESC LIMIT 1 FOR UPDATE`,
		tOptionTradingCalendarRows, m.table)
	var item TOptionTradingCalendar
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, calendarCode); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionTradingCalendarModel) FindOpenEndedApprovedForUpdate(
	ctx context.Context, tenantId int64, calendarCode string,
) (*TOptionTradingCalendar, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND calendar_code=? AND status=? AND effective_until=0
ORDER BY version DESC LIMIT 1 FOR UPDATE`, tOptionTradingCalendarRows, m.table)
	var item TOptionTradingCalendar
	if err := m.QueryRowNoCacheCtx(
		ctx, &item, query, tenantId, calendarCode,
		int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_APPROVED),
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionTradingCalendarModel) FindEffective(
	ctx context.Context, tenantId int64, calendarCode string, now int64,
) (*TOptionTradingCalendar, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND calendar_code=? AND status IN (?,?)
  AND effective_from<=? AND (effective_until=0 OR effective_until>?)
ORDER BY version DESC LIMIT 2`, tOptionTradingCalendarRows, m.table)
	var items []*TOptionTradingCalendar
	if err := m.QueryRowsNoCacheCtx(
		ctx, &items, query, tenantId, calendarCode,
		int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_APPROVED),
		int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_SUPERSEDED),
		now, now,
	); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	if len(items) != 1 {
		return nil, fmt.Errorf("ambiguous effective option trading calendar")
	}
	return items[0], nil
}

func (m *defaultTOptionTradingCalendarModel) FindPage(
	ctx context.Context, filter OptionTradingCalendarPageFilter, cursor, limit int64,
) ([]*TOptionTradingCalendar, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqString("calendar_code", filter.CalendarCode)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...,
	); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionTradingCalendarRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionTradingCalendar
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionTradingCalendarModel returns a model for the database table.
func NewTOptionTradingCalendarModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionTradingCalendarModel {
	return &customTOptionTradingCalendarModel{
		defaultTOptionTradingCalendarModel: newTOptionTradingCalendarModel(conn, c, opts...),
	}
}
