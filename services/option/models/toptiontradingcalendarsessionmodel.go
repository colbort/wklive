package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionTradingCalendarSessionModel = (*customTOptionTradingCalendarSessionModel)(nil)

type (
	// TOptionTradingCalendarSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionTradingCalendarSessionModel.
	TOptionTradingCalendarSessionModel interface {
		tOptionTradingCalendarSessionModel
		FindByCalendar(ctx context.Context, tenantId, calendarId int64) ([]*TOptionTradingCalendarSession, error)
	}

	customTOptionTradingCalendarSessionModel struct {
		*defaultTOptionTradingCalendarSessionModel
	}
)

func (m *defaultTOptionTradingCalendarSessionModel) FindByCalendar(
	ctx context.Context, tenantId, calendarId int64,
) ([]*TOptionTradingCalendarSession, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND calendar_id=? ORDER BY weekday,open_second,id`,
		tOptionTradingCalendarSessionRows, m.table)
	var items []*TOptionTradingCalendarSession
	err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, calendarId)
	return items, err
}

// NewTOptionTradingCalendarSessionModel returns a model for the database table.
func NewTOptionTradingCalendarSessionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionTradingCalendarSessionModel {
	return &customTOptionTradingCalendarSessionModel{
		defaultTOptionTradingCalendarSessionModel: newTOptionTradingCalendarSessionModel(conn, c, opts...),
	}
}
