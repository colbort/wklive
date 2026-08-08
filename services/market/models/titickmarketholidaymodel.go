package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickMarketHolidayModel = (*customTItickMarketHolidayModel)(nil)

type (
	// TItickMarketHolidayModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickMarketHolidayModel.
	TItickMarketHolidayModel interface {
		tItickMarketHolidayModel
		UpsertByCalendarDate(ctx context.Context, data *TItickMarketHoliday) error
	}

	customTItickMarketHolidayModel struct {
		*defaultTItickMarketHolidayModel
	}
)

// NewTItickMarketHolidayModel returns a model for the database table.
func NewTItickMarketHolidayModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickMarketHolidayModel {
	return &customTItickMarketHolidayModel{
		defaultTItickMarketHolidayModel: newTItickMarketHolidayModel(conn, c, opts...),
	}
}

func (m *customTItickMarketHolidayModel) UpsertByCalendarDate(ctx context.Context, data *TItickMarketHoliday) error {
	query := fmt.Sprintf(`INSERT INTO %s (calendar_id,trade_date,day_type,open_time,close_time,remark)
		VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),day_type=VALUES(day_type),
		open_time=VALUES(open_time),close_time=VALUES(close_time),remark=VALUES(remark)`, m.table)
	result, err := m.ExecNoCacheCtx(ctx, query, data.CalendarId, data.TradeDate, data.DayType, data.OpenTime, data.CloseTime, data.Remark)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	data.Id = id
	indexKey := fmt.Sprintf("%s%v:%v", cacheTItickMarketHolidayCalendarIdTradeDatePrefix, data.CalendarId, data.TradeDate)
	idKey := fmt.Sprintf("%s%v", cacheTItickMarketHolidayIdPrefix, id)
	return m.DelCache(indexKey, idKey)
}
