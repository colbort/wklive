package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionTradingCalendarExceptionModel = (*customTOptionTradingCalendarExceptionModel)(nil)

type (
	// TOptionTradingCalendarExceptionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionTradingCalendarExceptionModel.
	TOptionTradingCalendarExceptionModel interface {
		tOptionTradingCalendarExceptionModel
		FindByCalendar(ctx context.Context, tenantId, calendarId int64) ([]*TOptionTradingCalendarException, error)
		FindActiveAt(ctx context.Context, tenantId, calendarId, now int64) ([]*TOptionTradingCalendarException, error)
	}

	customTOptionTradingCalendarExceptionModel struct {
		*defaultTOptionTradingCalendarExceptionModel
	}
)

func (m *defaultTOptionTradingCalendarExceptionModel) FindByCalendar(
	ctx context.Context, tenantId, calendarId int64,
) ([]*TOptionTradingCalendarException, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND calendar_id=? ORDER BY start_time,end_time,id`,
		tOptionTradingCalendarExceptionRows, m.table)
	var items []*TOptionTradingCalendarException
	err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, calendarId)
	return items, err
}

func (m *defaultTOptionTradingCalendarExceptionModel) FindActiveAt(
	ctx context.Context, tenantId, calendarId, now int64,
) ([]*TOptionTradingCalendarException, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND calendar_id=? AND start_time<=? AND end_time>?
ORDER BY exception_type,id`, tOptionTradingCalendarExceptionRows, m.table)
	var items []*TOptionTradingCalendarException
	err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, calendarId, now, now)
	return items, err
}

// NewTOptionTradingCalendarExceptionModel returns a model for the database table.
func NewTOptionTradingCalendarExceptionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionTradingCalendarExceptionModel {
	return &customTOptionTradingCalendarExceptionModel{
		defaultTOptionTradingCalendarExceptionModel: newTOptionTradingCalendarExceptionModel(conn, c, opts...),
	}
}
